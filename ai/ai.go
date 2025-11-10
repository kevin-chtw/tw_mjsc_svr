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
	validActions := extractValidActionsFromOperates(state.Operates)

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

	for _, action := range validActions {
		switch action {
		case mahjong.OperateDiscard:
			// 打牌决策
			for i := range 34 {
				tile := mahjong.FromIndex(i)
				// 检查牌是否属于缺门花色
				isLackColor := tile.Color() == state.PlayerLacks[state.CurrentSeat]
				// 优先打缺门花色的牌
				bonus := float32(0)
				if isLackColor {
					bonus = 0.5 // 给缺门花色牌额外加分
				}
				if state.Hand[tile] > 0 && (qValues[i]+bonus) > bestDecision.QValue {
					reason := "最优打牌选择"
					if isLackColor {
						reason = "优先打缺门花色"
					}
					bestDecision = Decision{
						Operate:    mahjong.OperateDiscard,
						Tile:       tile,
						QValue:     qValues[i] + bonus,
						Reason:     reason,
						Confidence: sigmoid(qValues[i] + bonus),
					}
				}
			}

		case mahjong.OperateHu:
			// 胡牌决策
			if state.Operates != nil && state.Operates.HasOperate(mahjong.OperateHu) {
				// 使用AI计算的Q值来决定胡牌置信度
				huTile := state.LastTile
				if huTile == 0 && state.SelfTurn {
					// 如果是自摸，选择手牌中Q值最高的牌
					for i := range 34 {
						if state.Hand[mahjong.FromIndex(i)] > 0 && qValues[i] > qValues[huTile] {
							huTile = mahjong.FromIndex(i)
						}
					}
				}
				return Decision{
					Operate:    mahjong.OperateHu,
					Tile:       huTile,
					QValue:     qValues[huTile],
					Reason:     "胡牌优先",
					Confidence: sigmoid(qValues[huTile]),
				}
			}

		case mahjong.OperatePon:
			// 碰牌决策
			if state.Operates != nil && state.Operates.HasOperate(mahjong.OperatePon) {
				bestPeng, qValue := pickBestPeng(ai, state.LastTile, state)
				if qValue > bestDecision.QValue {
					bestDecision = Decision{
						Operate:    mahjong.OperatePon,
						Tile:       bestPeng,
						QValue:     qValue,
						Reason:     "高价值碰牌",
						Confidence: sigmoid(qValue),
					}
				}
			}

		case mahjong.OperateKon:
			// 杠牌决策
			if state.Operates != nil && state.Operates.HasOperate(mahjong.OperateKon) {
				bestGang, qValue := pickBestGang(ai, state.LastTile, state)
				if qValue > bestDecision.QValue {
					bestDecision = Decision{
						Operate:    mahjong.OperateKon,
						Tile:       bestGang,
						QValue:     qValue,
						Reason:     "高价值杠牌",
						Confidence: sigmoid(qValue),
					}
				}
			}
		}
	}

	// 5. 学习更新
	if ai.learnable && bestDecision.Operate == mahjong.OperateDiscard {
		ai.onlineUpdate(obs, mahjong.ToIndex(bestDecision.Tile), qValues, feat)
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

// extractValidActionsFromOperates 从 mahjong.Operates 中提取可行动作
func extractValidActionsFromOperates(operates *mahjong.Operates) []int {
	var validActions []int

	// 这里需要根据 mahjong.Operates 的实际结构来提取可行动作
	// 简化版本：假设总是可以打牌，并根据 operates 的内容判断其他动作
	validActions = append(validActions, mahjong.OperateDiscard)

	// 检查是否有胡牌操作
	if operates != nil && operates.HasOperate(mahjong.OperateHu) {
		validActions = append(validActions, mahjong.OperateHu)
	}

	// 检查是否有碰牌操作
	if operates != nil && operates.HasOperate(mahjong.OperatePon) {
		validActions = append(validActions, mahjong.OperatePon)
	}

	// 检查是否有杠牌操作
	if operates != nil && operates.HasOperate(mahjong.OperateKon) {
		validActions = append(validActions, mahjong.OperateKon)
	}

	return validActions
}

// sigmoid 辅助函数，将Q值转换为置信度
func sigmoid(x float32) float32 {
	return 1.0 / (1.0 + float32(math.Exp(float64(-x))))
}

// onlineUpdate 边打边学（私有）
func (ai *RichAI) onlineUpdate(obs []float32, act int, q []float32, feat *RichFeature) {
	// 即时 reward：牌效 + 向听 + 危险度
	oldHand := feat.Hand
	newHand := oldHand
	newHand[act]--
	effi := EffiReward(oldHand, newHand, 0) // lack 占位
	danger := 0.0
	for _, d := range DangerMask(feat, 0) {
		if !d {
			danger += 0.02
		}
	}
	r := effi - float32(danger)

	// Double-DQN：主网络选动作，target 网络评价值
	nextQ := ai.target.Forward(obs) // 简化仍用同 obs
	nextAct := 0
	for i := range 34 {
		if q[i] > q[nextAct] {
			nextAct = i
		}
	}
	maxNext := nextQ[nextAct]

	target := make([]float32, 34)
	copy(target, q)
	target[act] = r + float32(0.95)*maxNext

	// PER 存储
	tdErr := target[act] - q[act]
	ai.per.Add(obs, target, float64(math.Abs(float64(tdErr))))

	// 训练 batch=32
	if ai.per.pq.Len() >= batchSize {
		states, targets := ai.per.Sample(batchSize)
		ai.mu.Lock()
		loss := ai.net.Train(states, targets)
		ai.mu.Unlock()
		if ai.steps%1000 == 0 {
			log.Printf("step %d loss %.4f", ai.steps, loss)
		}
	}
	// 软更新 target
	if ai.steps%500 == 0 {
		ai.net.SoftUpdate(ai.target)
	}
	ai.steps++
}

// GameEndUpdate 终局奖励更新（新增）
func (ai *RichAI) GameEndUpdate(finalObs []float32, isWin bool, finalScore float32) {
	if !ai.learnable {
		return
	}

	// 终局奖励：赢+1，输-1，根据分数调整
	endReward := float32(-1.0)
	if isWin {
		endReward = 1.0
	}
	// 根据最终得分调整奖励
	endReward += finalScore * 0.1 // 每10分额外+1奖励

	// 获取当前状态的Q值
	ai.mu.RLock()
	qValues := ai.net.Forward(finalObs)
	ai.mu.RUnlock()

	// 创建目标Q值，所有动作都给予终局奖励
	target := make([]float32, 34)
	for i := range target {
		target[i] = endReward
	}

	// 存储到经验回放池
	tdErr := endReward - qValues[0] // 使用第一个Q值作为参考
	ai.per.Add(finalObs, target, float64(math.Abs(float64(tdErr))))

	log.Printf("Game end update: isWin=%v, score=%.1f, reward=%.2f", isWin, finalScore, endReward)
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
