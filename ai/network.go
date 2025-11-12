package ai

import (
	"fmt"
	"strconv"

	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

const (
	inputDim  = 604
	outputDim = 137
)

// DQNet 网络结构定义
type DQNet struct {
	g          *gorgonia.ExprGraph
	input      *gorgonia.Node
	qVals      *gorgonia.Node
	targetQ    *gorgonia.Node
	loss       *gorgonia.Node
	learnables gorgonia.Nodes
	solver     gorgonia.Solver
	vm         gorgonia.VM
}

// 创建 DQNet 实例
func NewDQNet() *DQNet {
	g := gorgonia.NewGraph()
	input := gorgonia.NewMatrix(g, tensor.Float32, gorgonia.WithShape(1, inputDim), gorgonia.WithName("input"))
	qVals := buildResNet18(g, input, outputDim)

	targetQ := gorgonia.NewMatrix(g, tensor.Float32, gorgonia.WithShape(1, outputDim), gorgonia.WithName("targetQ"))
	diff := gorgonia.Must(gorgonia.Sub(qVals, targetQ))
	loss := gorgonia.Must(gorgonia.Mean(gorgonia.Must(gorgonia.Square(diff))))

	// 收集所有可训练参数（conv/fc 的 W 与 B）
	var learnables gorgonia.Nodes
	for _, n := range g.AllNodes() {
		if n.Op() == nil && len(n.Shape()) > 0 &&
			(n.Name() == "" || n.Name()[len(n.Name())-1] == 'w' || n.Name()[len(n.Name())-1] == 'b') {
			learnables = append(learnables, n)
		}
	}
	solver := gorgonia.NewAdamSolver(gorgonia.WithLearnRate(1e-3))

	return &DQNet{
		g:          g,
		input:      input,
		qVals:      qVals,
		targetQ:    targetQ,
		loss:       loss,
		learnables: learnables,
		solver:     solver,
		vm:         gorgonia.NewTapeMachine(g),
	}
}

// 用于模型前向传播
func (net *DQNet) Forward(x []float32) []float32 {
	inTensor := tensor.New(tensor.WithShape(1, inputDim), tensor.WithBacking(x))
	if err := gorgonia.Let(net.input, inTensor); err != nil {
		logger.Log.Infof("Let failed: %v", err)
		return make([]float32, outputDim)
	}
	if err := gorgonia.Let(net.targetQ, tensor.New(
		tensor.WithShape(1, outputDim),
		tensor.WithBacking(make([]float32, outputDim)))); err != nil {
		logger.Log.Infof("Let targetQ patch failed: %v", err)
		return make([]float32, outputDim)
	}
	net.vm.Reset()
	if err := net.vm.RunAll(); err != nil {
		logger.Log.Infof("RunAll failed: %v", err)
		return make([]float32, outputDim)
	}
	return net.qVals.Value().Data().([]float32)
}

// 用于模型训练
func (net *DQNet) Train(states [][]float32, targets [][]float32) float32 {
	batch := len(states)
	if batch == 0 {
		return 0
	}
	var (
		totalLoss float32
		success   int
	)
	for i := 0; i < batch; i++ {
		if err := gorgonia.Let(net.input, tensor.New(
			tensor.WithShape(1, inputDim),
			tensor.WithBacking(states[i]),
		)); err != nil {
			logger.Log.Errorf("Let input failed: %v", err)
			continue
		}
		if err := gorgonia.Let(net.targetQ, tensor.New(
			tensor.WithShape(1, outputDim),
			tensor.WithBacking(targets[i]),
		)); err != nil {
			logger.Log.Errorf("Let targetQ failed: %v", err)
			continue
		}
		net.vm.Reset()
		if err := net.vm.RunAll(); err != nil {
			logger.Log.Errorf("VM Run failed: %v", err)
			continue
		}
		if lossVal, ok := net.loss.Value().Data().(float32); ok {
			totalLoss += lossVal
			success++
		}
	}
	if success == 0 {
		return 0
	}
	return totalLoss / float32(success)
}

// ResNet-18 架构实现
func buildResNet18(g *gorgonia.ExprGraph, input *gorgonia.Node, classNum int) *gorgonia.Node {
	// 输入：(B, 604) → reshape 成 (B,1,604,1) 给 4D 卷积用
	h := gorgonia.Must(gorgonia.Reshape(input, []int{-1, 1, inputDim, 1})) // 这里是 reshape

	// 调试输出形状
	logger.Log.Infof("After Reshape, shape: %v", h.Shape())

	h = conv2d(g, h, 64, 1, 1, 1, 1, 0, 0, "conv1") // 修改卷积核大小为 1x1
	h = relu(h)

	// 4 stages（各 2 个 BasicBlock，第一个 block stride=2 宽度方向下采样）
	h = makeLayer(g, h, 64, 2, 1, "layer1")
	h = makeLayer(g, h, 128, 2, 2, "layer2")
	h = makeLayer(g, h, 256, 2, 2, "layer3")
	h = makeLayer(g, h, 512, 2, 2, "layer4")

	// 调试输出每一层卷积后的形状
	logger.Log.Infof("After layer4, shape: %v", h.Shape())

	// GlobalAveragePool
	h = gorgonia.Must(gorgonia.Mean(h, 2))
	h = gorgonia.Must(gorgonia.Mean(h, 2))

	// 调试输出池化后的形状
	logger.Log.Infof("After GlobalAveragePool, shape: %v", h.Shape())

	// 全连接层，输出 137 个类别
	wFC := gorgonia.NewMatrix(g, tensor.Float32, gorgonia.WithShape(512, classNum),
		gorgonia.WithName("fc_w"), gorgonia.WithInit(gorgonia.GlorotN(1.0)))
	return gorgonia.Must(gorgonia.Mul(h, wFC))
}

func conv2d(g *gorgonia.ExprGraph, x *gorgonia.Node, outCh, kH, kW, strideH, strideW, padH, padW int, name string) *gorgonia.Node {
	inCh := int(x.Shape()[1])

	filter := gorgonia.NewTensor(g, tensor.Float32, 4,
		gorgonia.WithShape(outCh, inCh, kH, kW),
		gorgonia.WithName(name+"_w"),
		gorgonia.WithInit(gorgonia.GlorotN(1.0)))

	result := gorgonia.Must(gorgonia.Conv2d(x, filter,
		tensor.Shape{kH, kW},
		[]int{padH, padW},
		[]int{strideH, strideW},
		[]int{1, 1}))

	logger.Log.Infof("After %s, shape: %v", name, result.Shape())
	return result
}

// ReLU 激活函数
func relu(x *gorgonia.Node) *gorgonia.Node {
	return gorgonia.Must(gorgonia.Rectify(x))
}

func basicBlock(g *gorgonia.ExprGraph, x *gorgonia.Node, outCh, strideW, strideH int, name string) *gorgonia.Node {
	// 主路径的卷积操作
	h := conv2d(g, x, outCh, 3, 1, strideW, 1, 1, 0, name+"_conv1")
	fmt.Println("Shape after conv1: ", h.Shape())
	h = relu(h)

	// 第二个卷积操作
	h = conv2d(g, h, outCh, 3, 1, 1, 1, 1, 0, name+"_conv2")
	fmt.Println("Shape after conv2: ", h.Shape())

	// 处理 shortcut
	var short *gorgonia.Node
	if strideW != 1 || x.Shape()[1] != outCh {
		short = conv2d(g, x, outCh, 1, 1, strideW, 1, 0, 0, name+"_shortcut")
		fmt.Println("Shape after shortcut conv: ", short.Shape())
	} else {
		short = x
	}
	// 确保shortcut和主路径形状匹配
	if short.Shape()[2] != h.Shape()[2] || short.Shape()[3] != h.Shape()[3] {
		short = gorgonia.Must(gorgonia.Reshape(short, h.Shape()))
	}

	// 如果 shortcut 和主路径形状不匹配，调整通道数
	if short.Shape()[1] != h.Shape()[1] {
		short = conv2d(g, short, h.Shape()[1], 1, 1, 1, 1, 0, 0, name+"_shortcut_conv")
		fmt.Println("Shape after shortcut conv adjustment: ", short.Shape())
	}

	// 最终加和
	output := gorgonia.Must(gorgonia.Add(h, short))
	fmt.Println("Shape after add: ", output.Shape())

	return output
}

// 生成多层残差块
func makeLayer(g *gorgonia.ExprGraph, x *gorgonia.Node, outCh int, blocks int, strideW int, name string) *gorgonia.Node {
	h := basicBlock(g, x, outCh, strideW, strideW, name+"_blk1")
	for i := 1; i < blocks; i++ {
		h = basicBlock(g, h, outCh, 1, 1, name+"_blk"+strconv.Itoa(i+1))
	}

	// 打印层的输出形状
	logger.Log.Infof("After %s, shape: %v", name, h.Shape())

	return h
}
