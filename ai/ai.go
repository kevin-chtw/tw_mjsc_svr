package ai

import (
	"encoding/gob"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sync"

	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

var inst *RichAI
var once sync.Once
var globalLearnable bool // 全局训练模式标志

type RichAI struct {
	net       *DQNet
	per       *PER
	mu        sync.RWMutex
	count     int
	trainStep int     // 训练步数，用于学习率衰减
	epsilon   float32 // ε-greedy 探索率
}

// SetTrainingMode 设置全局训练模式（在程序启动时调用）
func SetTrainingMode(enable bool) {
	globalLearnable = enable
}

func GetRichAI() *RichAI {
	once.Do(func() {
		inst = &RichAI{
			net:       NewDQNet(),
			per:       NewPER(20000),
			count:     0,
			trainStep: 0,
			epsilon:   0.2, // 初始探索率20%
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

	// 先统一计算一次 Q 值，避免重复计算
	feat := state.ToRichFeature()
	obs := feat.ToVector()

	ai.mu.RLock()
	qValues := ai.net.Forward(obs)
	epsilon := ai.epsilon
	ai.mu.RUnlock()

	var candidates []*Decision // 收集所有候选动作
	state.SelfTurn = state.Operates.HasOperate(mahjong.OperateDiscard)

	if state.Operates.HasOperate(mahjong.OperateDiscard) {
		candidates = append(candidates, ai.addDiscards(state, qValues)...)
	}
	if state.Operates.HasOperate(mahjong.OperateHu) {
		if d := ai.hu(state, qValues); d != nil {
			candidates = append(candidates, d)
		}
	}
	if state.Operates.HasOperate(mahjong.OperatePon) {
		if d := ai.pon(state, qValues); d != nil {
			candidates = append(candidates, d)
		}
	}
	if state.Operates.HasOperate(mahjong.OperateKon) {
		candidates = append(candidates, ai.addKons(state, qValues)...)
	}
	if state.Operates.HasOperate(mahjong.OperatePass) {
		if d := ai.pass(state, qValues); d != nil {
			candidates = append(candidates, d)
		}
	}

	if len(candidates) == 0 {
		logger.Log.Errorf("==========================No candidates, epsilon=%.3f================", epsilon)
		return nil
	}

	// ε-greedy 策略：在所有候选动作中统一应用
	var bestD *Decision
	if globalLearnable && len(candidates) > 1 && rand.Float32() < epsilon {
		// 探索：从所有候选动作中随机选择
		bestD = candidates[rand.Intn(len(candidates))]
	} else {
		// 利用：选择Q值最高的动作
		for _, d := range candidates {
			if bestD == nil || d.QValue > bestD.QValue {
				bestD = d
			}
		}
	}

	if globalLearnable && bestD != nil && bestD.Operate != int(mahjong.OperateNone) {
		logger.Log.Infof("==========================RecordDecision  %v (epsilon=%.3f, candidates=%d)================", bestD, epsilon, len(candidates))
		state.RecordDecision(bestD.Operate, bestD.Tile)
	}

	return bestD
}

func (ai *RichAI) addTiles(state *GameState, qValues []float32, isLack bool) []*Decision {
	var result []*Decision
	lackColor := state.PlayerLacks[state.CurrentSeat]

	for tile, count := range state.Hand {
		if count <= 0 {
			continue
		}
		if isLack != (tile.Color() == lackColor) {
			continue
		}

		value := ai.evaluateState(qValues, mahjong.OperateDiscard, tile)
		result = append(result, &Decision{
			Operate: mahjong.OperateDiscard,
			Tile:    tile,
			QValue:  value,
		})
	}
	return result
}

func (ai *RichAI) addDiscards(state *GameState, qValues []float32) []*Decision {
	// 优先收集缺门牌，如果没有再收集其他牌
	result := ai.addTiles(state, qValues, true)
	if len(result) == 0 {
		result = ai.addTiles(state, qValues, false)
	}
	return result
}

func (ai *RichAI) hu(state *GameState, qValues []float32) *Decision {
	value := ai.evaluateState(qValues, mahjong.OperateHu, state.LastTile)
	return &Decision{
		Operate: mahjong.OperateHu,
		Tile:    state.LastTile,
		QValue:  value,
	}
}

func (ai *RichAI) pon(state *GameState, qValues []float32) *Decision {
	value := ai.evaluateState(qValues, mahjong.OperatePon, state.LastTile)
	return &Decision{
		Operate: mahjong.OperatePon,
		Tile:    state.LastTile,
		QValue:  value,
	}
}

func (ai *RichAI) addKons(state *GameState, qValues []float32) []*Decision {
	var result []*Decision
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

	for _, t := range tiles {
		value := ai.evaluateState(qValues, mahjong.OperateKon, t)
		result = append(result, &Decision{
			Operate: mahjong.OperateKon,
			Tile:    t,
			QValue:  value,
		})
	}
	return result
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

// 从 obs 中提取可行动作掩码并计算掩码后的最大 Q 值
// obs 向量中，Operates 的 5 维 one-hot 位于索引 [600, 605)
// 动作索引区间定义见 actionIndex:
//
//	Discard: [0,34), Pon: [34,68), Kon: [68,102), Hu: [102,136), Pass: [136,137)
func maxMaskedQFromObs(q []float32, obs []float32) float32 {
	const operatesOffset = 600
	const operatesDim = 5
	if len(obs) < operatesOffset+operatesDim {
		max := float32(-1e9)
		for _, v := range q {
			if v > max {
				max = v
			}
		}
		return max
	}
	ops := obs[operatesOffset : operatesOffset+operatesDim]
	ranges := make([][2]int, 0, 5)
	// Discard
	if ops[0] > 0.5 {
		ranges = append(ranges, [2]int{0, 34})
	}
	// Hu
	if ops[1] > 0.5 {
		ranges = append(ranges, [2]int{102, 136})
	}
	// Pon
	if ops[2] > 0.5 {
		ranges = append(ranges, [2]int{34, 68})
	}
	// Kon
	if ops[3] > 0.5 {
		ranges = append(ranges, [2]int{68, 102})
	}
	// Pass
	if ops[4] > 0.5 {
		ranges = append(ranges, [2]int{136, 137})
	}
	if len(ranges) == 0 {
		max := float32(-1e9)
		for _, v := range q {
			if v > max {
				max = v
			}
		}
		return max
	}
	max := float32(-1e9)
	for _, r := range ranges {
		for i := r[0]; i < r[1]; i++ {
			if q[i] > max {
				max = q[i]
			}
		}
	}
	return max
}

// GameEndUpdate 带指标的训练更新（支持 reward shaping）
func (ai *RichAI) GameEndUpdate(finalState *GameState) {
	if !globalLearnable {
		return
	}

	ai.mu.Lock()
	defer ai.mu.Unlock()

	γ := float32(0.99)
	historyLen := len(finalState.DecisionHistory)
	nextMaxQ := float32(0.0)

	// 从 GameState 中获取终局信息
	finalScore := finalState.FinalScore
	isLiuJu := finalState.IsLiuJu
	decisionSteps := len(finalState.DecisionHistory)

	// 判断是否胡牌及番数
	isHu := false
	var huMulti int64
	for _, huSeat := range finalState.HuPlayers {
		if huSeat == finalState.CurrentSeat {
			isHu = true
			if multi, ok := finalState.HuMultis[huSeat]; ok {
				huMulti = multi
			}
			break
		}
	}

	// 计算增强奖励（reward shaping）
	shapedReward := finalScore

	// 1. 胡牌奖励：基础分 + 番数加成
	if isHu {
		// 胡牌基础奖励
		shapedReward += 10.0
		// 番数奖励（鼓励做大番）
		shapedReward += float32(huMulti) * 2.0
		// 快速胡牌奖励（步数越少奖励越高）
		if decisionSteps > 0 && decisionSteps < 50 {
			speedBonus := (50.0 - float32(decisionSteps)) / 10.0
			shapedReward += speedBonus
		}
	} else if isLiuJu {
		// 2. 流局惩罚（只惩罚未胡牌的AI，鼓励主动胡牌）
		shapedReward -= 5.0
	}

	for i := historyLen - 1; i >= 0; i-- {
		decisionRec := finalState.DecisionHistory[i]

		currQ := ai.net.Forward(decisionRec.Obs)

		// 先计算当前状态的 max Q 值（带可行动作掩码），用于下一个（更早的）决策
		maxCurrQ := maxMaskedQFromObs(currQ, decisionRec.Obs)

		// 计算 TD 目标：最后一步使用 shapedReward，其他步骤使用 γ * nextMaxQ
		var tdTarget float32
		if i == historyLen-1 {
			// 最后一步：使用温和的归一化，保留更多信号强度
			// tanh(x/5) 比 tanh(x/20) 保留更多值的范围
			tdTarget = float32(math.Tanh(float64(shapedReward) / 5.0))
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

		// 直接将目标动作的目标值设置为 tdTarget（不进行二次"软更新"）
		target[actionIdx] = tdTarget
		tdErr := float32(math.Abs(float64(tdTarget - currQ[actionIdx])))
		ai.per.Add(decisionRec.Obs, target, float64(tdErr))
	}

	states, targets := ai.per.Sample()
	loss := ai.net.Train(states, targets)
	ai.trainStep++

	// epsilon 衰减：从 0.2 衰减到 0.05，在 10000 步后稳定
	if ai.trainStep < 10000 {
		ai.epsilon = 0.2 - (0.15 * float32(ai.trainStep) / 10000.0)
	} else {
		ai.epsilon = 0.05
	}

	// 增强日志：记录关键指标
	logger.Log.Infof("==================================GameEndUpdate, loss=%.6f, trainStep=%d, epsilon=%.3f, isHu=%v, multi=%d, liuju=%v, steps=%d, shapedReward=%.2f",
		loss, ai.trainStep, ai.epsilon, isHu, huMulti, isLiuJu, decisionSteps, shapedReward)
}

// SaveWeights 把网络权重落盘到文件
func (ai *RichAI) SaveWeights(path string) error {
	if !globalLearnable {
		return nil // 非训练模式不保存权重
	}
	if ai.trainStep%40 != 0 {
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
