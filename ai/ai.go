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
	Operate int          `json:"operate"` // mahjong.Operate 类型
	Tile    mahjong.Tile `json:"tile"`    // 牌值
	QValue  float32      `json:"q_value"` // 决策Q值
}

// Step 统一决策入口 - 外部传入可行操作和局面
func (ai *RichAI) Step(state *GameState) *Decision {
	d := &Decision{
		QValue: -1e9,
	}

	if state.Operates.HasOperate(mahjong.OperateDiscard) {
		d = ai.discard(state, d)
	}

	if state.Operates.HasOperate(mahjong.OperateHu) {
		d = ai.hu(state, d)
	}
	if state.Operates.HasOperate(mahjong.OperatePon) {
		d = ai.pon(state, d)
	}
	if state.Operates.HasOperate(mahjong.OperateKon) {
		d = ai.kon(state, d)
	}
	if state.Operates.HasOperate(mahjong.OperatePass) {
		d = ai.pass(state, d)
	}

	if ai.learnable && d.Operate != int(mahjong.OperateNone) {
		state.RecordAction(d.Operate, d.Tile)
	}

	return d
}

func (ai *RichAI) discard(state *GameState, d *Decision) *Decision {
	for i := range 34 {
		tile := mahjong.FromIndex(i)
		if state.Hand[tile] <= 0 {
			continue
		}

		// 模拟打牌
		state.Hand[tile]--
		state.PlayerMelds[state.CurrentSeat][tile]++

		value := ai.evaluateState(state)
		// 恢复状态
		state.Hand[tile]++
		state.PlayerMelds[state.CurrentSeat][tile]--

		if value > d.QValue {
			d = &Decision{
				Operate: mahjong.OperateDiscard,
				Tile:    tile,
				QValue:  value,
			}
		}
	}
	return d
}

func (ai *RichAI) hu(state *GameState, d *Decision) *Decision {
	state.HuTiles[state.CurrentSeat] = state.LastTile
	state.HuPlayers = append(state.HuPlayers, state.CurrentSeat)

	value := ai.evaluateState(state)
	// 恢复状态
	state.HuTiles[state.CurrentSeat] = mahjong.TileNull
	state.HuPlayers = state.HuPlayers[:len(state.HuPlayers)-1]

	if value > d.QValue {
		d = &Decision{
			Operate: mahjong.OperateHu,
			Tile:    state.LastTile,
			QValue:  value,
		}
	}
	return d
}

func (ai *RichAI) pon(state *GameState, d *Decision) *Decision {
	state.PonTiles[state.CurrentSeat] = append(state.PonTiles[state.CurrentSeat], state.LastTile)
	state.Hand[state.LastTile] -= 2

	value := ai.evaluateState(state)
	// 恢复状态
	state.PonTiles[state.CurrentSeat] = state.PonTiles[state.CurrentSeat][:len(state.PonTiles[state.CurrentSeat])-1]
	state.Hand[state.LastTile] += 2

	if value > d.QValue {
		d = &Decision{
			Operate: mahjong.OperatePon,
			Tile:    state.LastTile,
			QValue:  value,
		}
	}
	return d
}

func (ai *RichAI) kon(state *GameState, d *Decision) *Decision {
	state.KonTiles[state.CurrentSeat] = append(state.KonTiles[state.CurrentSeat], state.LastTile)
	count := state.Hand[state.LastTile]
	state.Hand[state.LastTile] = 0

	value := ai.evaluateState(state)

	// 恢复状态
	state.KonTiles[state.CurrentSeat] = state.KonTiles[state.CurrentSeat][:len(state.KonTiles[state.CurrentSeat])-1]
	state.Hand[state.LastTile] += count

	if value > d.QValue {
		d = &Decision{
			Operate: mahjong.OperateKon,
			Tile:    state.LastTile,
			QValue:  value,
		}
	}
	return d
}

func (ai *RichAI) pass(state *GameState, d *Decision) *Decision {
	value := ai.evaluateState(state)
	if value > d.QValue {
		d = &Decision{
			Operate: mahjong.OperatePass,
			Tile:    mahjong.TileNull,
			QValue:  value,
		}
	}
	return d
}

func (ai *RichAI) evaluateState(state *GameState) float32 {
	feat := state.ToRichFeature()
	obs := feat.ToVector()

	ai.mu.RLock()
	defer ai.mu.RUnlock()

	// 只用 target 网络评估局面价值
	qValues := ai.target.Forward(obs)

	maxQ := float32(-1e9)
	for _, q := range qValues {
		if q > maxQ {
			maxQ = q
		}
	}
	return maxQ
}

// GameEndUpdate 终局奖励更新（基于整个局面）
// GameEndUpdate 终局奖励更新（dense reward + 单动作更新 + target网络）
func (ai *RichAI) GameEndUpdate(finalState *GameState, isWin bool, finalScore float32) {
	if !ai.learnable {
		return
	}

	ai.mu.Lock()
	defer ai.mu.Unlock()

	baseReward := 1.0 + finalScore*0.1

	// 2. 用 target 网络评估终局价值（避免高估）
	finalFeat := finalState.ToRichFeature()
	finalObs := finalFeat.ToVector()
	finalQValues := ai.target.Forward(finalObs) // ← 用 target 网络

	// 3. 逆序遍历历史，逐步回传奖励（蒙特卡洛回溯）
	for i := len(finalState.ActionHistory) - 1; i >= 0; i-- {
		rec := &finalState.ActionHistory[i]

		// 3.1 中间动作奖励
		stepReward := baseReward
		switch rec.Operate {
		case mahjong.OperateHu:
			stepReward += 0.1
		case mahjong.OperateKon:
			stepReward += 0.05
		case mahjong.OperatePon:
			stepReward += 0.02
		}

		// 3.2 当前状态 Q 值（主网络）
		currQ := ai.net.Forward(rec.Obs)

		// 3.3 TD 目标：r + γ * max_a' Q_target(s', a')
		γ := float32(0.99)
		maxNextQ := float32(-1e9)
		for _, q := range finalQValues {
			if q > maxNextQ {
				maxNextQ = q
			}
		}
		tdTarget := stepReward + γ*maxNextQ

		// 3.4 只更新实际执行的动作
		target := make([]float32, 137)
		copy(target, currQ)                                        // 其余动作保持不变
		target[actionIndex(rec.Operate, rec.TileIndex)] = tdTarget // 仅更新本动作

		// 3.5 TD 误差
		tdErr := float32(math.Abs(float64(tdTarget - currQ[rec.TileIndex])))

		// 3.6 存入 PER
		ai.per.Add(rec.Obs, target, float64(tdErr))
	}

	// 4. 立即训练一次
	batchSize := 256
	if ai.per.Len() >= batchSize {
		states, targets := ai.per.Sample(batchSize)
		loss := ai.net.Train(states, targets)
		log.Printf("GameEndUpdate: trained %d samples, loss=%.6f", batchSize, loss)
		ai.SaveWeights()
	}
}
