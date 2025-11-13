package ai

import (
	"encoding/gob"
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

var inst *RichAI
var once sync.Once

type RichAI struct {
	net       *DQNet
	per       *PER
	learnable bool
	mu        sync.RWMutex
	count     int
	trainStep int // 训练步数，用于学习率衰减
}

func GetRichAI(learnable bool) *RichAI {
	once.Do(func() {
		inst = &RichAI{
			net:       NewDQNet(),
			per:       NewPER(20000),
			learnable: learnable,
			count:     0,
			trainStep: 0,
		}
		loadWeights(inst.net, "tw_mjsc_svr.gob")
	})
	return inst
}

func loadWeights(net *DQNet, path string) error {
	var w map[string][]float32
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := gob.NewDecoder(f).Decode(&w); err != nil {
		return err
	}
	for _, n := range net.learnables {
		weightData, exists := w[n.Name()]
		if !exists {
			logger.Log.Warnf("Weight %s not found in file, using random initialization", n.Name())
			continue
		}
		expectedSize := 1
		for _, dim := range n.Shape() {
			expectedSize *= dim
		}
		if len(weightData) != expectedSize {
			logger.Log.Warnf("Weight %s shape mismatch. Expected %d elements, got %d. Using random initialization", n.Name(), expectedSize, len(weightData))
			continue
		}
		if err := gorgonia.Let(n, tensor.New(tensor.WithShape(n.Shape()...), tensor.WithBacking(weightData))); err != nil {
			logger.Log.Warnf("Failed to load weight %s: %v. Using random initialization", n.Name(), err)
			continue
		}
	}
	return nil
}

// Decision 统一决策结果和决策记录
type Decision struct {
	Operate int          `json:"operate"`       // mahjong.Operate 类型
	Tile    mahjong.Tile `json:"tile"`          // 牌值
	QValue  float32      `json:"q_value"`       // 决策Q值（用于选择最佳决策）
	Obs     []float32    `json:"obs,omitempty"` // 操作时的状态特征（用于训练，可选）
}

// Step 统一决策入口 - 外部传入可行操作和局面
func (ai *RichAI) Step(state *GameState) *Decision {
	var bestD *Decision

	// 根据请求中是否有出牌操作设置 SelfTurn
	if state.Operates != nil {
		state.SelfTurn = state.Operates.HasOperate(mahjong.OperateDiscard)
	}

	// 先统一计算一次 Q 值，避免重复计算
	feat := state.ToRichFeature()
	obs := feat.ToVector()

	ai.mu.RLock()
	qValues := ai.net.Forward(obs)
	ai.mu.RUnlock()

	// 辅助函数：更新最佳决策
	updateBest := func(d *Decision) {
		if d != nil && (bestD == nil || d.QValue > bestD.QValue) {
			bestD = d
		}
	}

	if state.Operates.HasOperate(mahjong.OperateDiscard) {
		updateBest(ai.discard(state, qValues))
	}

	if state.Operates.HasOperate(mahjong.OperateHu) {
		updateBest(ai.hu(state, qValues))
	}
	if state.Operates.HasOperate(mahjong.OperatePon) {
		updateBest(ai.pon(state, qValues))
	}
	if state.Operates.HasOperate(mahjong.OperateKon) {
		updateBest(ai.kon(state, qValues))
	}
	if state.Operates.HasOperate(mahjong.OperatePass) {
		updateBest(ai.pass(state, qValues))
	}

	if ai.learnable && bestD != nil && bestD.Operate != int(mahjong.OperateNone) {
		logger.Log.Infof("==========================RecordDecision  %v================", bestD)
		state.RecordDecision(bestD.Operate, bestD.Tile)
	}

	return bestD
}

func (ai *RichAI) findBestDiscard(state *GameState, qValues []float32, isLack bool) *Decision {
	var bestD *Decision
	lackColor := state.PlayerLacks[state.CurrentSeat]

	for tile, count := range state.Hand {
		if count <= 0 {
			continue
		}
		if isLack != (tile.Color() == lackColor) {
			continue
		}

		value := ai.evaluateState(qValues, mahjong.OperateDiscard, tile)
		if state.CallData != nil {
			if _, ok := state.CallData[tile.ToInt32()]; ok {
				value += 0.5
			}
		}

		if bestD == nil || value > bestD.QValue {
			bestD = &Decision{
				Operate: mahjong.OperateDiscard,
				Tile:    tile,
				QValue:  value,
			}
		}
	}

	return bestD
}

func (ai *RichAI) discard(state *GameState, qValues []float32) *Decision {
	if bestLackD := ai.findBestDiscard(state, qValues, true); bestLackD != nil {
		return bestLackD
	}
	return ai.findBestDiscard(state, qValues, false)
}

func (ai *RichAI) hu(state *GameState, qValues []float32) *Decision {
	value := ai.evaluateState(qValues, mahjong.OperateHu, state.LastTile) + 2
	return &Decision{
		Operate: mahjong.OperateHu,
		Tile:    state.LastTile,
		QValue:  value,
	}
}

func (ai *RichAI) pon(state *GameState, qValues []float32) *Decision {
	value := ai.evaluateState(qValues, mahjong.OperatePon, state.LastTile) + 0.5
	return &Decision{
		Operate: mahjong.OperatePon,
		Tile:    state.LastTile,
		QValue:  value,
	}
}

func (ai *RichAI) kon(state *GameState, qValues []float32) *Decision {
	tiles := make([]mahjong.Tile, 0)
	if state.SelfTurn {
		for _, t := range state.PonTiles[state.CurrentSeat] {
			if state.Hand[t] == 1 {
				tiles = append(tiles, t)
			}
		}
		for t, count := range state.Hand {
			if count == 4 {
				tiles = append(tiles, t)
			}
		}
	} else {
		tiles = append(tiles, state.LastTile)
	}

	var bestD *Decision
	for _, t := range tiles {
		value := ai.evaluateState(qValues, mahjong.OperateKon, t)
		if bestD == nil || value > bestD.QValue {
			bestD = &Decision{
				Operate: mahjong.OperateKon,
				Tile:    t,
				QValue:  value,
			}
		}
	}
	return bestD
}

func (ai *RichAI) pass(_ *GameState, qValues []float32) *Decision {
	value := ai.evaluateState(qValues, mahjong.OperatePass, mahjong.TileNull)
	return &Decision{
		Operate: mahjong.OperatePass,
		Tile:    mahjong.TileNull,
		QValue:  value,
	}
}

func getActionIndex(operate int, tile mahjong.Tile) (int, error) {
	if operate == mahjong.OperatePass {
		return 136, nil
	}
	if !tile.IsValid() {
		return 0, fmt.Errorf("invalid tile: operate=%d, tile=%d", operate, tile)
	}
	tileIdx := mahjong.ToIndex(tile)
	actionIdx := actionIndex(operate, tileIdx)
	if actionIdx < 0 || actionIdx >= outputDim {
		return 0, fmt.Errorf("invalid action index: %d", actionIdx)
	}
	return actionIdx, nil
}

func (ai *RichAI) evaluateState(qValues []float32, operate int, tile mahjong.Tile) float32 {
	actionIdx, err := getActionIndex(operate, tile)
	if err != nil {
		logger.Log.Warnf("getActionIndex failed: %v, returning 0", err)
		return 0
	}

	return qValues[actionIdx]
}

func (ai *RichAI) GameEndUpdate(finalState *GameState, finalScore float32) {
	if !ai.learnable {
		return
	}

	ai.mu.Lock()
	defer ai.mu.Unlock()

	γ := float32(0.99)
	historyLen := len(finalState.DecisionHistory)
	nextMaxQ := float32(0.0)

	for i := historyLen - 1; i >= 0; i-- {
		decisionRec := finalState.DecisionHistory[i]

		currQ := ai.net.Forward(decisionRec.Obs)

		// 先计算当前状态的 max Q 值，用于下一个（更早的）决策
		maxCurrQ := float32(-1e9)
		for _, q := range currQ {
			if q > maxCurrQ {
				maxCurrQ = q
			}
		}

		// 计算 TD 目标：最后一步使用 finalScore，其他步骤使用 γ * nextMaxQ
		var tdTarget float32
		if i == historyLen-1 {
			// 最后一步：tdTarget = finalScore（最终状态之后没有未来）
			tdTarget = 1.0 + finalScore*0.1
		} else {
			// 其他步骤：tdTarget = γ * nextMaxQ（使用下一个状态的 max Q 值）
			tdTarget = γ * nextMaxQ
		}

		// 更新 nextMaxQ 为当前状态的 max Q 值，供下一个（更早的）决策使用
		nextMaxQ = maxCurrQ

		target := make([]float32, 137)
		copy(target, currQ)

		actionIdx, err := getActionIndex(decisionRec.Operate, decisionRec.Tile)
		if err != nil {
			logger.Log.Warnf("getActionIndex failed: %v, skipping", err)
			continue
		}

		learningRate := float32(1.0)
		if ai.trainStep > 0 {
			decayFactor := float32(math.Pow(0.9, float64(ai.trainStep)/1000.0))
			learningRate = float32(math.Max(0.1, float64(1.0*decayFactor)))
		}
		oldValue := currQ[actionIdx]
		target[actionIdx] = (1-learningRate)*oldValue + learningRate*tdTarget
		tdErr := float32(math.Abs(float64(tdTarget - currQ[actionIdx])))
		ai.per.Add(decisionRec.Obs, target, float64(tdErr))
	}

	states, targets := ai.per.Sample()
	loss := ai.net.Train(states, targets)
	ai.trainStep++
	logger.Log.Infof("==================================GameEndUpdate, loss=%.6f, trainStep=%d", loss, ai.trainStep)
}

// saveWeights 把网络权重落盘到文件
func (ai *RichAI) SaveWeights(path string) error {
	ai.count++
	if ai.count%4 != 0 {
		return nil
	}
	logger.Log.Infof("==================================SaveWeights================")
	ai.mu.Lock()
	defer ai.mu.Unlock()
	w := make(map[string][]float32) // 导出字段
	for _, n := range ai.net.learnables {
		w[n.Name()] = n.Value().Data().([]float32)
	}
	f, err := os.Create(path)
	if err != nil {
		logger.Log.Errorf("==================================SaveWeights failed: %v", err)
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(w)
}
