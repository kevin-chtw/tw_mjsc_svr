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

	if ai.learnable && d.Operate != 0 {
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

		state.Hand[tile]--
		state.PlayerMelds[state.CurrentSeat][tile]++
		defer func() {
			state.Hand[tile]++
			state.PlayerMelds[state.CurrentSeat][tile]--
		}()

		currentQ := ai.currentScore(state)
		if currentQ > d.QValue {
			d = &Decision{
				Operate: mahjong.OperateDiscard,
				Tile:    tile,
				QValue:  currentQ,
			}
		}
	}
	return d
}

func (ai *RichAI) hu(state *GameState, d *Decision) *Decision {
	state.HuTiles[state.CurrentSeat] = state.LastTile
	state.HuPlayers = append(state.HuPlayers, state.CurrentSeat)
	defer func() {
		state.HuTiles[state.CurrentSeat] = mahjong.TileNull
		state.HuPlayers = state.HuPlayers[:len(state.HuPlayers)-1]
	}()

	currentQ := ai.currentScore(state)
	if currentQ > d.QValue {
		d = &Decision{
			Operate: mahjong.OperateHu,
			Tile:    state.LastTile,
			QValue:  currentQ,
		}
	}
	return d
}

func (ai *RichAI) pon(state *GameState, d *Decision) *Decision {
	state.PonTiles[state.CurrentSeat] = append(state.PonTiles[state.CurrentSeat], state.LastTile)
	state.Hand[state.LastTile] -= 2
	defer func() {
		state.PonTiles[state.CurrentSeat] = state.PonTiles[state.CurrentSeat][:len(state.PonTiles[state.CurrentSeat])-1]
		state.Hand[state.LastTile] += 2
	}()

	currentQ := ai.currentScore(state)
	if currentQ > d.QValue {
		d = &Decision{
			Operate: mahjong.OperatePon,
			Tile:    state.LastTile,
			QValue:  currentQ,
		}
	}
	return d
}

func (ai *RichAI) kon(state *GameState, d *Decision) *Decision {
	state.KonTiles[state.CurrentSeat] = append(state.KonTiles[state.CurrentSeat], state.LastTile)
	count := state.Hand[state.LastTile]
	state.Hand[state.LastTile] = 0
	defer func() {
		state.KonTiles[state.CurrentSeat] = state.KonTiles[state.CurrentSeat][:len(state.KonTiles[state.CurrentSeat])-1]
		state.Hand[state.LastTile] += count
	}()

	currentQ := ai.currentScore(state)
	if currentQ > d.QValue {
		d = &Decision{
			Operate: mahjong.OperateKon,
			Tile:    state.LastTile,
			QValue:  currentQ,
		}
	}
	return d
}

func (ai *RichAI) pass(state *GameState, d *Decision) *Decision {
	currentQ := ai.currentScore(state)
	if currentQ > d.QValue {
		d = &Decision{
			Operate: mahjong.OperateHu,
			Tile:    mahjong.TileNull,
			QValue:  currentQ,
		}
	}
	return d
}

// currentScore 返回当前局面的整体评分（平均Q值）
func (ai *RichAI) currentScore(state *GameState) float32 {
	qValues := ai.currentQ(state)
	total := float32(0)
	count := 0

	for _, q := range qValues {
		total += q
		count++
	}

	if count > 0 {
		return total / float32(count)
	}
	return 0
}

func (ai *RichAI) currentQ(state *GameState) []float32 {
	newFeat := state.ToRichFeature()
	newObs := newFeat.ToVector()

	ai.mu.RLock()
	defer ai.mu.RUnlock()

	return ai.net.Forward(newObs)
}

// GameEndUpdate 终局奖励更新（修改为累计奖励）
func (ai *RichAI) GameEndUpdate(finalState *GameState, isWin bool, finalScore float32) {
	if !ai.learnable && !isWin {
		return
	}

	ai.mu.Lock()
	defer ai.mu.Unlock()

	// 计算所有历史操作的奖励并累加
	for _, record := range finalState.ActionHistory {
		// 无论输赢，都将每步结果更新到经验回放库
		obs := record.Feature.ToVector()
		target := make([]float32, 34)
		for i := range target {
			target[i] = 1.0
		}
		ai.per.Add(obs, target, 1.0)
	}

	// 根据最终得分调整奖励
	totalReward := 1.0 + finalScore*0.1

	// 转换最终状态为特征向量
	finalFeat := finalState.ToRichFeature()
	finalObs := finalFeat.ToVector()
	// 获取当前状态的Q值
	qValues := ai.net.Forward(finalObs)

	// 创建目标Q值，基于当前Q值但用累计奖励更新实际执行的动作
	target := make([]float32, 34)
	copy(target, qValues) // 首先复制当前Q值

	// 对于历史记录中的每个操作，用累计奖励更新对应的target
	for _, record := range finalState.ActionHistory {
		if record.TileIndex >= 0 && record.TileIndex < 34 {
			target[record.TileIndex] = totalReward
		}
	}

	// 计算平均TD误差用于优先级排序
	tdErr := float32(0)
	count := 0
	for i := range target {
		if target[i] != qValues[i] { // 只计算被更新的动作
			tdErr += float32(math.Abs(float64(target[i] - qValues[i])))
			count++
		}
	}
	if count > 0 {
		tdErr /= float32(count)
	}

	// 存储到经验回放池
	ai.per.Add(finalObs, target, float64(tdErr))
}

// NotifyGameResult 通知游戏结果（公开接口）
func (ai *RichAI) NotifyGameResult(finalState *GameState, isWin bool, finalScore float32) {
	if !ai.learnable {
		return
	}

	// 调用终局奖励更新
	ai.GameEndUpdate(finalState, isWin, finalScore)

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
