package ai

import (
	"encoding/gob"
	"log"
	"math"
	"os"
	"sync"

	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
)

var inst *RichAI
var once sync.Once

type RichAI struct {
	net       *DQNet
	target    *DQNet
	per       *PER
	learnable bool
	mu        sync.RWMutex
	steps     int
	history   []ActionRecord // 新增：操作历史记录
}

// ActionRecord 记录每一步操作
type ActionRecord struct {
	Operate   int
	TileIndex int
	Obs       []float32
	QValues   []float32
	Feature   *RichFeature
	TimeStep  int
}

func GetRichAI(learnable bool) *RichAI {
	once.Do(func() {
		inst = &RichAI{
			net:       NewDQNet(),
			target:    NewDQNet(),
			per:       NewPER(20000),
			learnable: learnable,
		}
		LoadWeights(inst.net, "rich_dqn.gob")
		LoadWeights(inst.target, "rich_dqn.gob")
	})
	return inst
}

// StepInput / StepOutput
type StepInput struct {
	RichFeature RichFeature
	Mask        [34]bool
}
type StepOutput struct {
	Action int       `json:"action"`
	Q      []float32 `json:"q"`
}

// Decision 统一决策结果
type Decision struct {
	Operate    int          `json:"operate"`    // mahjong.Operate 类型
	Tile       mahjong.Tile `json:"tile"`       // 牌值
	QValue     float32      `json:"q_value"`    // 决策Q值
	Reason     string       `json:"reason"`     // 决策原因
	Confidence float32      `json:"confidence"` // 置信度
}

// Step 统一决策入口 - 外部传入可行操作和局面
func (ai *RichAI) Step(state *GameState) Decision {
	// 1. 从 mahjong.Operates 中提取可行动作
	validActions := extractperates(state.Operates)

	// 2. 预处理: 转换状态为特征向量
	feat := state.ToRichFeature()
	obs := feat.ToVector()

	// 3. 获取网络输出
	ai.mu.RLock()
	qValues := ai.net.Forward(obs)
	ai.mu.RUnlock()

	// 4. 根据可行动作进行决策
	var bestDecision Decision
	bestDecision.QValue = -1e9

	// 预计算所有可能的决策
	var decisions []Decision

	for _, action := range validActions {
		switch action {
		case mahjong.OperateDiscard:
			// 打牌决策 - 找出所有可能的打牌选择
			for i := range 34 {
				tile := mahjong.FromIndex(i)
				if state.Hand[tile] <= 0 {
					continue
				}

				isLackColor := tile.Color() == state.PlayerLacks[state.CurrentSeat]
				bonus := float32(0)
				if isLackColor {
					bonus = 0.5
				}

				currentQ := qValues[i] + bonus
				decisions = append(decisions, Decision{
					Operate:    mahjong.OperateDiscard,
					Tile:       tile,
					QValue:     currentQ,
					Reason:     "打牌选择",
					Confidence: sigmoid(currentQ),
				})
			}

		case mahjong.OperateHu:
			// 胡牌决策
			if state.Operates != nil && state.Operates.HasOperate(mahjong.OperateHu) {
				huTile := state.LastTile
				if huTile == 0 && state.SelfTurn {
					// 自摸时选择手牌中Q值最高的牌
					maxQ := float32(-1e9)
					for i := range 34 {
						tile := mahjong.FromIndex(i)
						if state.Hand[tile] > 0 && qValues[i] > maxQ {
							maxQ = qValues[i]
							huTile = tile
						}
					}
				}

				if huTile != 0 {
					decisions = append(decisions, Decision{
						Operate:    mahjong.OperateHu,
						Tile:       huTile,
						QValue:     qValues[mahjong.ToIndex(huTile)],
						Reason:     "胡牌优先",
						Confidence: sigmoid(qValues[mahjong.ToIndex(huTile)]),
					})
				}
			}

		case mahjong.OperatePon:
			// 碰牌决策
			if state.Operates != nil && state.Operates.HasOperate(mahjong.OperatePon) {
				bestPeng, qValue := pickBestPeng(ai, state.LastTile, state)
				decisions = append(decisions, Decision{
					Operate:    mahjong.OperatePon,
					Tile:       bestPeng,
					QValue:     qValue,
					Reason:     "高价值碰牌",
					Confidence: sigmoid(qValue),
				})
			}

		case mahjong.OperateKon:
			// 杠牌决策
			if state.Operates != nil && state.Operates.HasOperate(mahjong.OperateKon) {
				bestGang, qValue := pickBestGang(ai, state.LastTile, state)
				decisions = append(decisions, Decision{
					Operate:    mahjong.OperateKon,
					Tile:       bestGang,
					QValue:     qValue,
					Reason:     "高价值杠牌",
					Confidence: sigmoid(qValue),
				})
			}

		case mahjong.OperatePass:
			// 过牌决策
			passQValue := calculatePassQValue(state, qValues)
			decisions = append(decisions, Decision{
				Operate:    mahjong.OperatePass,
				Tile:       0,
				QValue:     passQValue,
				Reason:     "主动选择过牌",
				Confidence: sigmoid(passQValue),
			})
		}
	}

	// 在所有可能的决策中选择Q值最高的
	for _, decision := range decisions {
		if decision.QValue > bestDecision.QValue {
			bestDecision = decision
		}
	}

	// 5. 记录操作历史用于终局奖励计算
	if ai.learnable && bestDecision.Operate != 0 {
		ai.recordActionHistory(bestDecision.Operate, mahjong.ToIndex(bestDecision.Tile), feat)
	}

	return bestDecision
}

// pickBestPeng 选择最优碰牌动作
func pickBestPeng(ai *RichAI, tile mahjong.Tile, state *GameState) (mahjong.Tile, float32) {
	// 转换状态为特征向量
	feat := state.ToRichFeature()
	obs := feat.ToVector()

	// 获取网络输出
	ai.mu.RLock()
	qValues := ai.net.Forward(obs)
	ai.mu.RUnlock()

	// 返回当前牌和对应的Q值
	return tile, qValues[tile]
}

// calculatePassQValue 计算"过"操作的Q值
func calculatePassQValue(state *GameState, qValues []float32) float32 {
	// 基础Q值为所有可能动作Q值的平均值
	totalQ := float32(0)
	count := 0

	// 计算所有可能打牌动作的Q值平均值
	for i := range 34 {
		tile := mahjong.FromIndex(i)
		if state.Hand[tile] > 0 {
			totalQ += qValues[i]
			count++
		}
	}

	if count > 0 {
		// "过"操作的Q值基于平均Q值，但稍微降低
		return totalQ/float32(count) - 0.1
	}

	// 如果没有可打的牌，返回一个保守的Q值
	return -0.5
}

// pickBestGang 选择最优杠牌动作
func pickBestGang(ai *RichAI, tile mahjong.Tile, state *GameState) (mahjong.Tile, float32) {
	// 转换状态为特征向量
	feat := state.ToRichFeature()
	obs := feat.ToVector()

	// 获取网络输出
	ai.mu.RLock()
	qValues := ai.net.Forward(obs)
	ai.mu.RUnlock()

	// 返回当前牌和对应的Q值
	return tile, qValues[tile]
}

// extractperates 从 mahjong.Operates 中提取可行动作
func extractperates(operates *mahjong.Operates) []int {
	var validActions []int

	if operates.HasOperate(mahjong.OperateDiscard) {
		validActions = append(validActions, mahjong.OperateDiscard)
	}

	if operates.HasOperate(mahjong.OperateHu) {
		validActions = append(validActions, mahjong.OperateHu)
	}

	if operates.HasOperate(mahjong.OperatePon) {
		validActions = append(validActions, mahjong.OperatePon)
	}

	if operates.HasOperate(mahjong.OperateKon) {
		validActions = append(validActions, mahjong.OperateKon)
	}

	if operates.HasOperate(mahjong.OperatePass) {
		validActions = append(validActions, mahjong.OperatePass)
	}
	return validActions
}

// sigmoid 辅助函数，将Q值转换为置信度
func sigmoid(x float32) float32 {
	return 1.0 / (1.0 + float32(math.Exp(float64(-x))))
}

// recordActionHistory 记录操作历史用于终局奖励计算
func (ai *RichAI) recordActionHistory(operate int, tileIndex int, feat *RichFeature) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	obs := feat.ToVector()
	qValues := ai.net.Forward(obs)

	record := ActionRecord{
		Operate:   operate,
		TileIndex: tileIndex,
		Obs:       obs,
		QValues:   qValues,
		Feature:   feat,
		TimeStep:  ai.steps,
	}
	ai.history = append(ai.history, record)
}

// calculateStepReward 计算单步操作的奖励
func (ai *RichAI) calculateStepReward(record ActionRecord) float32 {
	var r float32
	switch record.Operate {
	case mahjong.OperateDiscard:
		// 打牌reward：牌效 + 危险度
		oldHand := record.Feature.Hand
		newHand := oldHand
		newHand[record.TileIndex]--
		effi := EffiReward(oldHand, newHand, 0) // lack 占位
		danger := 0.0
		for _, d := range DangerMask(record.Feature, 0) {
			if !d {
				danger += 0.02
			}
		}
		r = effi - float32(danger)

	case mahjong.OperateHu:
		// 胡牌reward：固定高奖励
		r = 1.0

	case mahjong.OperatePon, mahjong.OperateKon:
		// 碰/杠reward：基于牌的价值
		r = float32(0.5 + 0.1*float64(record.QValues[record.TileIndex]))

	case mahjong.OperatePass:
		// 过牌reward：保守的负奖励
		r = -0.3
	}
	return r
}

// GameEndUpdate 终局奖励更新（修改为累计奖励）
func (ai *RichAI) GameEndUpdate(finalObs []float32, isWin bool, finalScore float32) {
	if !ai.learnable {
		return
	}

	ai.mu.Lock()
	defer ai.mu.Unlock()

	// 计算累计奖励：终局奖励 + 所有历史操作奖励
	totalReward := float32(0.0)

	// 终局基础奖励
	if isWin {
		totalReward += 1.0
	} else {
		totalReward -= 1.0
	}
	// 根据最终得分调整奖励
	totalReward += finalScore * 0.1

	// 计算所有历史操作的奖励并累加
	for _, record := range ai.history {
		stepReward := ai.calculateStepReward(record)
		totalReward += stepReward
		log.Printf("Step %d reward: %.2f (operate: %d, tile: %d)",
			record.TimeStep, stepReward, record.Operate, record.TileIndex)
	}

	// 获取当前状态的Q值
	qValues := ai.net.Forward(finalObs)

	// 创建目标Q值，所有动作都给予累计奖励
	target := make([]float32, 34)
	for i := range target {
		target[i] = totalReward
	}

	// 存储到经验回放池
	tdErr := totalReward - qValues[0]
	ai.per.Add(finalObs, target, float64(math.Abs(float64(tdErr))))

	// 清空历史记录
	ai.history = nil

	log.Printf("Game end update: isWin=%v, score=%.1f, total_reward=%.2f, steps=%d",
		isWin, finalScore, totalReward, len(ai.history))
}

// NotifyGameResult 通知游戏结果（公开接口）
func (ai *RichAI) NotifyGameResult(finalState *GameState, isWin bool, finalScore float32) {
	if !ai.learnable {
		return
	}

	// 转换最终状态为特征向量
	finalFeat := finalState.ToRichFeature()
	finalObs := finalFeat.ToVector()

	// 调用终局奖励更新
	ai.GameEndUpdate(finalObs, isWin, finalScore)

	log.Printf("Game result notified: isWin=%v, score=%.1f", isWin, finalScore)
}

// SaveWeights 落盘
func (ai *RichAI) SaveWeights() error {
	ai.mu.RLock()
	defer ai.mu.RUnlock()
	return saveWeights(ai.net, "rich_dqn.gob")
}

// saveWeights 把网络权重落盘到文件
func saveWeights(net *DQNet, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(net)
}

func LoadWeights(net *DQNet, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err // 文件不存在时返回错误，调用方可以忽略
	}
	defer f.Close()
	return gob.NewDecoder(f).Decode(net)
}
