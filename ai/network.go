package ai

import (
	"strconv"

	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

const (
	inputDim     = 605 + HistoryDim // 605 + 2580 = 3185（基础特征 + 历史操作序列）
	outputDim    = 137
	expandedH    = 57 // 57×56 = 3192，最接近 3185
	expandedW    = 56
	expandedSize = expandedH * expandedW // 3192
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
	inferVM    gorgonia.VM // 专用于推理的VM，避免计算loss
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
	solver := gorgonia.NewAdamSolver(gorgonia.WithLearnRate(3e-3)) // 增加学习率以加快收敛

	// 【关键修复】构建梯度图，计算所有可学习参数的梯度
	if _, err := gorgonia.Grad(loss, learnables...); err != nil {
		logger.Log.Fatalf("Failed to build gradient graph: %v", err)
	}

	vm := gorgonia.NewTapeMachine(g)
	return &DQNet{
		g:          g,
		input:      input,
		qVals:      qVals,
		targetQ:    targetQ,
		loss:       loss,
		learnables: learnables,
		solver:     solver,
		vm:         vm,
		inferVM:    vm, // 使用同一个VM，避免状态不一致
	}
}

// 用于模型前向传播（推理模式）
func (net *DQNet) Forward(x []float32) []float32 {
	inTensor := tensor.New(tensor.WithShape(1, inputDim), tensor.WithBacking(x))
	if err := gorgonia.Let(net.input, inTensor); err != nil {
		logger.Log.Infof("Let failed: %v", err)
		return make([]float32, outputDim)
	}
	// 【优化】推理时给 targetQ 赋值相同维度的零向量，避免未初始化错误
	if err := gorgonia.Let(net.targetQ, tensor.New(
		tensor.WithShape(1, outputDim),
		tensor.WithBacking(make([]float32, outputDim)))); err != nil {
		logger.Log.Infof("Let targetQ patch failed: %v", err)
		return make([]float32, outputDim)
	}
	net.inferVM.Reset()
	if err := net.inferVM.RunAll(); err != nil {
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
	// 【修复】语法错误：for i := range batch 无法遍历int，改为标准for循环
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

		// 【关键修复】执行反向传播和权重更新
		// 将 gorgonia.Nodes 转换为 []gorgonia.ValueGrad
		valueGrads := make([]gorgonia.ValueGrad, len(net.learnables))
		for idx, node := range net.learnables {
			valueGrads[idx] = node
		}
		if err := net.solver.Step(valueGrads); err != nil {
			logger.Log.Errorf("Solver step failed: %v", err)
		}
	}
	if success == 0 {
		return 0
	}
	return totalLoss / float32(success)
}

func buildResNet18(g *gorgonia.ExprGraph, input *gorgonia.Node, classNum int) *gorgonia.Node {
	zeroPad := gorgonia.NewMatrix(g, tensor.Float32,
		gorgonia.WithShape(1, expandedSize-inputDim),
		gorgonia.WithName("zero_pad"),
		gorgonia.WithInit(gorgonia.Zeroes()))

	expanded := gorgonia.Must(gorgonia.Concat(1, input, zeroPad))

	x := gorgonia.Must(gorgonia.Reshape(expanded, []int{-1, 1, expandedH, expandedW}))

	x = conv2d(g, x, 64, 7, 7, 2, 2, 3, 3, "conv1")
	x = batchNorm(g, x, 64, "bn1")
	x = gorgonia.Must(gorgonia.Rectify(x))

	x = maxPool2d(x, 3, 3, 2, 2, 1, 1)

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
	shape := x.Shape()
	if len(shape) >= 4 && shape[2] == 1 && shape[3] == 1 {
		return gorgonia.Must(gorgonia.Reshape(x, []int{-1, shape[1]}))
	}
	out, err := gorgonia.Mean(x, 2, 3)
	if err != nil {
		logger.Log.Warnf("Mean failed, using Reshape instead: %v", err)
		return gorgonia.Must(gorgonia.Reshape(x, []int{-1, shape[1]}))
	}
	return out
}
