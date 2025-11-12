package ai

import (
	"strconv"

	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

const (
	inputDim     = 604 + HistoryDim // 604 + 4300 = 4904（基础特征 + 历史操作序列）
	outputDim    = 137
	expandedH    = 71 // 71×70 = 4970，最接近 4904
	expandedW    = 70
	expandedSize = expandedH * expandedW // 4970
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

// ResNet-18 架构实现（标准版，适配 604 维输入）
func buildResNet18(g *gorgonia.ExprGraph, input *gorgonia.Node, classNum int) *gorgonia.Node {
	// 输入：(B, 604) → 扩展到 (B, expandedSize) → reshape 成 (B, 1, expandedH, expandedW)
	// 604 → 608 (19×32)，用零填充
	// 注意：由于训练时 batch size 总是 1，使用固定形状 (1, expandedSize-inputDim)
	zeroPad := gorgonia.NewMatrix(g, tensor.Float32,
		gorgonia.WithShape(1, expandedSize-inputDim),
		gorgonia.WithName("zero_pad"),
		gorgonia.WithInit(gorgonia.Zeroes()))

	// 拼接：input + zeroPad → (B, expandedSize)
	expanded := gorgonia.Must(gorgonia.Concat(1, input, zeroPad))

	// Reshape 成 (B, 1, expandedH, expandedW) 用于 2D 卷积
	x := gorgonia.Must(gorgonia.Reshape(expanded, []int{-1, 1, expandedH, expandedW}))

	// 标准 ResNet-18 第一层：7×7 卷积，stride=2，padding=3
	x = conv2d(g, x, 64, 7, 7, 2, 2, 3, 3, "conv1")
	x = batchNorm(g, x, 64, "bn1")
	x = gorgonia.Must(gorgonia.Rectify(x))

	// MaxPool: 3×3，stride=2，padding=1
	x = maxPool2d(x, 3, 3, 2, 2, 1, 1)

	// 4 个残差层，每个 2 个 block
	x = makeLayer(g, x, 64, 2, 1, "layer1")
	x = makeLayer(g, x, 128, 2, 2, "layer2")
	x = makeLayer(g, x, 256, 2, 2, "layer3")
	x = makeLayer(g, x, 512, 2, 2, "layer4")

	// Global Average Pooling
	x = globalAvgPool2d(x)
	x = gorgonia.Must(gorgonia.Reshape(x, []int{-1, 512}))

	// 全连接层
	wFC := gorgonia.NewMatrix(g, tensor.Float32, gorgonia.WithShape(512, classNum),
		gorgonia.WithName("fc_w"), gorgonia.WithInit(gorgonia.GlorotN(1.0)))

	return gorgonia.Must(gorgonia.Mul(x, wFC))
}

func conv2d(g *gorgonia.ExprGraph, x *gorgonia.Node, outCh, kH, kW, strideH, strideW, padH, padW int, name string) *gorgonia.Node {
	inCh := int(x.Shape()[1])

	filter := gorgonia.NewTensor(g, tensor.Float32, 4,
		gorgonia.WithShape(outCh, inCh, kH, kW),
		gorgonia.WithName(name+"_w"),
		gorgonia.WithInit(gorgonia.GlorotN(1.0)))

	return gorgonia.Must(gorgonia.Conv2d(x, filter,
		tensor.Shape{kH, kW},
		[]int{padH, padW},
		[]int{strideH, strideW},
		[]int{1, 1}))
}

func batchNorm(g *gorgonia.ExprGraph, x *gorgonia.Node, channels int, name string) *gorgonia.Node {
	scale := gorgonia.NewTensor(g, tensor.Float32, 4,
		gorgonia.WithShape(1, channels, 1, 1),
		gorgonia.WithName(name+"_scale_w"),
		gorgonia.WithInit(gorgonia.Ones()))
	bias := gorgonia.NewTensor(g, tensor.Float32, 4,
		gorgonia.WithShape(1, channels, 1, 1),
		gorgonia.WithName(name+"_bias_b"),
		gorgonia.WithInit(gorgonia.Zeroes()))

	out, _, _, _, err := gorgonia.BatchNorm(x, scale, bias, 0.9, 1e-5)
	if err != nil {
		panic(err)
	}
	return out
}

func maxPool2d(x *gorgonia.Node, kH, kW, strideH, strideW, padH, padW int) *gorgonia.Node {
	ret, err := gorgonia.MaxPool2D(x, tensor.Shape{kH, kW}, []int{padH, padW}, []int{strideH, strideW})
	if err != nil {
		panic(err)
	}
	return ret
}

func basicBlock(g *gorgonia.ExprGraph, x *gorgonia.Node, outCh int, stride int, name string) *gorgonia.Node {
	// 标准 ResNet-18：3×3 卷积
	h := conv2d(g, x, outCh, 3, 3, stride, stride, 1, 1, name+"_conv1")
	h = batchNorm(g, h, outCh, name+"_bn1")
	h = gorgonia.Must(gorgonia.Rectify(h))

	h = conv2d(g, h, outCh, 3, 3, 1, 1, 1, 1, name+"_conv2")
	h = batchNorm(g, h, outCh, name+"_bn2")

	var shortcut *gorgonia.Node
	if stride != 1 || int(x.Shape()[1]) != outCh {
		shortcut = conv2d(g, x, outCh, 1, 1, stride, stride, 0, 0, name+"_shortcut_conv")
		shortcut = batchNorm(g, shortcut, outCh, name+"_shortcut_bn")
	} else {
		shortcut = x
	}

	out := gorgonia.Must(gorgonia.Add(h, shortcut))
	return gorgonia.Must(gorgonia.Rectify(out))
}

func makeLayer(g *gorgonia.ExprGraph, x *gorgonia.Node, outCh int, blocks int, stride int, name string) *gorgonia.Node {
	h := basicBlock(g, x, outCh, stride, name+"_blk1")
	for i := 1; i < blocks; i++ {
		h = basicBlock(g, h, outCh, 1, name+"_blk"+strconv.Itoa(i+1))
	}
	logger.Log.Infof("After %s, shape: %v", name, h.Shape())
	return h
}

func globalAvgPool2d(x *gorgonia.Node) *gorgonia.Node {
	// 对于很小的空间维度（如 1×1），直接 Reshape 即可
	// 对于较大的空间维度，使用 Mean 操作
	shape := x.Shape()
	if len(shape) >= 4 && shape[2] == 1 && shape[3] == 1 {
		// 空间维度已经是 1×1，直接 Reshape
		return gorgonia.Must(gorgonia.Reshape(x, []int{-1, shape[1]}))
	}
	// 否则使用 Mean 操作（对空间维度求平均）
	out, err := gorgonia.Mean(x, 2, 3)
	if err != nil {
		// 如果 Mean 失败，尝试使用 Reshape（假设空间维度很小）
		logger.Log.Warnf("Mean failed, using Reshape instead: %v", err)
		return gorgonia.Must(gorgonia.Reshape(x, []int{-1, shape[1]}))
	}
	return out
}
