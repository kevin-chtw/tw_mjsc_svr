package ai

import (
	"fmt"
	"math"
	"time"

	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
)

var inst *RichAI
var globalLearnable bool

type RichAI struct{}

func SetTrainingMode(enable bool) {
	globalLearnable = enable
}

func IsTrainingMode() bool {
	return globalLearnable
}

func GetRichAI() *RichAI {
	if inst == nil {
		inst = &RichAI{}
	}
	return inst
}

type Decision struct {
	Operate int          `json:"operate"`
	Tile    mahjong.Tile `json:"tile"`
	QValue  float32      `json:"q_value"`
	Obs     []float32    `json:"obs,omitempty"`
}

// Step - 通过 HTTP 调用 Python AI 服务
func (ai *RichAI) Step(state *GameState) *Decision {
	start := time.Now()

	// 生成观察向量
	feat := state.ToRichFeature()
	obs := feat.ToVector()

	// 收集所有可行的动作
	var candidates []*Decision
	if state.Operates.HasOperate(mahjong.OperateDiscard) {
		candidates = append(candidates, ai.addDiscards(state)...)
	}
	if state.Operates.HasOperate(mahjong.OperateHu) {
		candidates = append(candidates, ai.hu(state))
	}
	if state.Operates.HasOperate(mahjong.OperatePon) {
		candidates = append(candidates, ai.pon(state))
	}
	if state.Operates.HasOperate(mahjong.OperateKon) {
		candidates = append(candidates, ai.addKons(state)...)
	}
	if state.Operates.HasOperate(mahjong.OperatePass) {
		candidates = append(candidates, ai.pass(state))
	}

	if len(candidates) == 0 {
		logger.Log.Errorf("No valid candidates")
		return nil
	}

	httpClient := GetHTTPAIClient()
	if httpClient == nil {
		logger.Log.Errorf("HTTP AI client not initialized, using fallback")
		return candidates[0]
	}

	// 发送GameState和候选动作给Python
	decision, err := httpClient.GetDecision(state, obs, candidates)
	if err != nil || decision == nil {
		logger.Log.Warnf("GetDecision failed: %v, using fallback to first candidate", err)
		return candidates[0]
	}

	// 验证决策是否在候选列表中
	found := false
	for _, cand := range candidates {
		if cand.Operate == decision.Operate && cand.Tile == decision.Tile {
			found = true
			break
		}
	}

	if !found {
		// 详细记录候选列表以便调试
		candStr := ""
		for i, cand := range candidates {
			if i > 0 {
				candStr += ", "
			}
			candStr += fmt.Sprintf("(%d,%d)", cand.Operate, cand.Tile)
			if i >= 10 {
				candStr += "..."
				break
			}
		}
		logger.Log.Warnf("AI returned invalid decision (operate=%d, tile=%d), not in candidates: [%s], using first candidate",
			decision.Operate, decision.Tile, candStr)
		return candidates[0]
	}

	// 记录决策用于训练
	if globalLearnable && decision.Operate != int(mahjong.OperateNone) {
		decision.Obs = obs
		state.RecordDecision(decision.Operate, decision.Tile, obs)
	}

	// 记录慢速决策
	totalTime := time.Since(start)
	if totalTime > 500*time.Millisecond {
		logger.Log.Warnf("Step slow: total=%v", totalTime)
	}

	return decision
}

func (ai *RichAI) addDiscards(state *GameState) []*Decision {
	decisions := make([]*Decision, 0)
	lackSuit := state.PlayerLacks[state.CurrentSeat]
	hasLack := lackSuit != 4 // SuitMax = 4，4表示没有缺门

	// 如果有缺门，优先收集缺门的牌
	if hasLack {
		for tile, count := range state.Hand {
			if count > 0 && tile.Color() == mahjong.EColor(lackSuit) {
				decisions = append(decisions, &Decision{
					Operate: int(mahjong.OperateDiscard),
					Tile:    tile,
				})
			}
		}
		// 如果有缺门的牌可以打，就只返回这些
		if len(decisions) > 0 {
			return decisions
		}
	}

	// 没有缺门，或者缺门已经打完了，可以打任何牌
	for tile, count := range state.Hand {
		if count > 0 {
			decisions = append(decisions, &Decision{
				Operate: int(mahjong.OperateDiscard),
				Tile:    tile,
			})
		}
	}
	return decisions
}

func (ai *RichAI) hu(state *GameState) *Decision {
	return &Decision{
		Operate: int(mahjong.OperateHu),
		Tile:    state.LastTile,
	}
}

func (ai *RichAI) pon(state *GameState) *Decision {
	return &Decision{
		Operate: int(mahjong.OperatePon),
		Tile:    state.LastTile,
	}
}

func (ai *RichAI) addKons(state *GameState) []*Decision {
	decisions := make([]*Decision, 0)
	if !state.Operates.HasOperate(mahjong.OperateKon) {
		return decisions
	}

	// 1. 暗杠：手里有4张相同的牌
	for tile, count := range state.Hand {
		if count == 4 {
			decisions = append(decisions, &Decision{
				Operate: int(mahjong.OperateKon),
				Tile:    tile,
			})
		}
	}

	// 2. 补杠：之前碰过的牌，手里又有1张
	if ponTiles, ok := state.PonTiles[state.CurrentSeat]; ok {
		for _, ponTile := range ponTiles {
			if state.Hand[ponTile] >= 1 {
				decisions = append(decisions, &Decision{
					Operate: int(mahjong.OperateKon),
					Tile:    ponTile,
				})
			}
		}
	}

	return decisions
}

func (ai *RichAI) pass(_ *GameState) *Decision {
	return &Decision{
		Operate: int(mahjong.OperatePass),
		Tile:    0,
	}
}

// QueueTraining - 游戏结束时发送训练数据到 Python
func (ai *RichAI) QueueTraining(finalState *GameState) {
	if !globalLearnable {
		return
	}

	decisionSteps := len(finalState.DecisionHistory)
	finalScore := finalState.FinalScore

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

	shapedReward := finalScore

	logger.Log.Warnf("QueueTraining: isHu=%v, multi=%d,  steps=%d, shapedReward=%.2f",
		isHu, huMulti, decisionSteps, shapedReward)

	httpClient := GetHTTPAIClient()
	if httpClient == nil {
		logger.Log.Warnf("HTTP AI client not initialized, skipping training")
		return
	}

	// 验证DecisionHistory有效性
	if len(finalState.DecisionHistory) == 0 {
		return
	}

	steps := make([]StepTransition, 0, len(finalState.DecisionHistory))
	validSteps := 0
	// 使用折扣因子分配奖励：越接近最终结果的决策，奖励越大
	// 例如：最后一步 100%，倒数第二步 80%，倒数第三步 64%...
	discountFactor := float32(0.97)

	totalReward := float32(0)
	for i := 0; i < len(finalState.DecisionHistory); i++ {
		rec := &finalState.DecisionHistory[i]

		// 计算折扣后的奖励：越早的决策，折扣越多
		stepsFromEnd := len(finalState.DecisionHistory) - 1 - i
		reward := shapedReward * float32(math.Pow(float64(discountFactor), float64(stepsFromEnd)))

		var nextState []float32
		done := false
		if i < len(finalState.DecisionHistory)-1 {
			nextState = finalState.DecisionHistory[i+1].Obs
		} else {
			done = true
		}
		totalReward += reward
		steps = append(steps, StepTransition{
			State:     rec.Obs,
			Operate:   rec.Operate,
			Tile:      mahjong.ToIndex(rec.Tile),
			Reward:    reward,
			NextState: nextState,
			Done:      done,
		})
		validSteps++
	}

	// 只有当有有效步骤时才发送
	if validSteps > 0 {
		episode := &Episode{
			Steps:        steps,
			ShapedReward: totalReward,
			IsHu:         isHu,
			HuMulti:      huMulti,
		}
		httpClient.ReportEpisode(episode)
		logger.Log.Infof("QueueTraining: sent %d valid steps out of %d total", validSteps, len(finalState.DecisionHistory))
	} else {
		logger.Log.Warnf("QueueTraining: no valid steps to send")
	}
}

func (ai *RichAI) SaveWeights(path string) error {
	return nil // Python manages weights
}
