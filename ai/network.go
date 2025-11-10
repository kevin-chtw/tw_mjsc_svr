package ai

import (
	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

const (
	inputDim  = 599 // 匹配 RichFeature.ToVector() 的实际输出长度
	outputDim = 34
	batchSize = 32
	tau       = 0.01
)

type DQNet struct {
	g          *gorgonia.ExprGraph
	input      *gorgonia.Node
	targetQ    *gorgonia.Node
	fcW        *gorgonia.Node
	qVals      *gorgonia.Node
	loss       *gorgonia.Node
	learnables []*gorgonia.Node
	vm         gorgonia.VM
	solver     gorgonia.Solver
}

func NewDQNet() *DQNet {
	g := gorgonia.NewGraph()

	// 构建图时同时放占位 tensor
	placeholderIn := tensor.New(tensor.WithShape(batchSize, inputDim), tensor.WithBacking(make([]float32, batchSize*inputDim)))
	placeholderTgt := tensor.New(tensor.WithShape(batchSize, outputDim), tensor.WithBacking(make([]float32, batchSize*outputDim)))

	input := gorgonia.NewMatrix(g, tensor.Float32, gorgonia.WithShape(batchSize, inputDim), gorgonia.WithValue(placeholderIn))
	targetQ := gorgonia.NewMatrix(g, tensor.Float32, gorgonia.WithShape(batchSize, outputDim), gorgonia.WithValue(placeholderTgt))

	// 单层权重矩阵
	fcW := gorgonia.NewMatrix(g, tensor.Float32, gorgonia.WithShape(inputDim, outputDim), gorgonia.WithInit(gorgonia.GlorotN(1.0)))

	// 简单线性变换
	qVals := gorgonia.Must(gorgonia.Mul(input, fcW))

	// loss: MSE
	diff := gorgonia.Must(gorgonia.Sub(qVals, targetQ))
	loss := gorgonia.Must(gorgonia.Mean(diff))

	// Adam
	solver := gorgonia.NewAdamSolver(
		gorgonia.WithLearnRate(0.001),
		gorgonia.WithBeta1(0.9),
		gorgonia.WithBeta2(0.999),
		gorgonia.WithEps(1e-8),
	)
	learnables := []*gorgonia.Node{fcW}
	vm := gorgonia.NewTapeMachine(g, gorgonia.BindDualValues(learnables...))

	return &DQNet{
		g:          g,
		input:      input,
		targetQ:    targetQ,
		fcW:        fcW,
		qVals:      qVals,
		loss:       loss,
		learnables: learnables,
		vm:         vm,
		solver:     solver,
	}
}

// Forward 单样本
func (net *DQNet) Forward(x []float32) []float32 {
	// 创建新 tensor
	inTensor := tensor.New(tensor.WithShape(1, inputDim), tensor.WithBacking(x))

	// 使用 gorgonia.Let 绑定输入值
	gorgonia.Let(net.input, inTensor)

	net.vm.Reset()
	if err := net.vm.RunAll(); err != nil {
		panic(err)
	}

	// 获取输出值
	output := net.qVals.Value()
	if output == nil {
		panic("qVals value is nil")
	}

	return output.Data().([]float32)
}

// Train 一次 batch
func (net *DQNet) Train(states [][]float32, targets [][]float32) float32 {
	// 使用 gorgonia.Let 绑定输入值
	gorgonia.Let(net.input, tensor.New(tensor.WithShape(batchSize, inputDim), tensor.WithBacking(flattenF32(states))))
	gorgonia.Let(net.targetQ, tensor.New(tensor.WithShape(batchSize, outputDim), tensor.WithBacking(flattenF32(targets))))

	net.vm.Reset()
	if err := net.vm.RunAll(); err != nil {
		panic(err)
	}
	lossVal := net.loss.Value().Data().(float32)
	net.vm.Reset()
	// 将 *Node 转换为 ValueGrad
	valueGrads := gorgonia.NodesToValueGrads(net.learnables)
	net.solver.Step(valueGrads)
	return lossVal
}

// SoftUpdate 复制权重 → target（τ=0.01）
func (net *DQNet) SoftUpdate(target *DQNet) {
	for i, src := range net.learnables {
		dst := target.learnables[i]
		srcVal := src.Value().Data().([]float32)
		dstVal := dst.Value().Data().([]float32)
		for j := range srcVal {
			dstVal[j] = tau*srcVal[j] + (1-tau)*dstVal[j]
		}
	}
}

func flattenF32(x [][]float32) []float32 {
	var out []float32
	for _, row := range x {
		out = append(out, row...)
	}
	return out
}
