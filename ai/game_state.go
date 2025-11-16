package ai

import (
	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/kevin-chtw/tw_proto/game/pbmj"
)

// ActionRecord 实现操作记录（根据 ack 记录的实际操作，用于生成特征）
type ActionRecord struct {
	Seat      int // 执行操作的玩家座位号（0-3）
	Operate   int // 操作类型
	TileIndex int // 牌索引
}

// GameState 小型状态快照
type GameState struct {
	Operates        *mahjong.Operates      // 可执行操作
	CurrentSeat     int                    // 当前玩家座位号（0-3）
	TotalTiles      int                    // 总牌张数
	LastTile        mahjong.Tile           // 最近打出的牌
	Hand            map[mahjong.Tile]int   // 手牌（牌->数量）
	PonTiles        map[int][]mahjong.Tile // 碰牌信息（玩家ID->碰的牌列表）
	KonTiles        map[int][]mahjong.Tile // 杠牌信息（玩家ID->杠的牌列表）
	HuTiles         map[int]mahjong.Tile   // 胡牌信息（玩家ID->胡的牌）
	HuMultis        map[int]int64          // 胡牌番数（玩家ID->番数）
	PlayerLacks     [4]mahjong.EColor      // 四个玩家的缺门花色
	HuPlayers       []int                  // 胡牌玩家ID列表
	DecisionHistory []Decision             // 决策历史记录（用于训练，不限制长度，不生成特征）
	ActionHistory   []ActionRecord         // 实现操作历史记录（用于生成特征，限制60条）
	CallData        map[int32]*pbmj.CallData
	// 终局统计信息
	FinalScore float32 // 最终得分（包含点炮惩罚）
}

func NewGameState() *GameState {
	return &GameState{
		TotalTiles:      136,
		Operates:        mahjong.NewOperates(int32(mahjong.OperateNone)),
		Hand:            make(map[mahjong.Tile]int),
		PonTiles:        make(map[int][]mahjong.Tile),
		KonTiles:        make(map[int][]mahjong.Tile),
		HuTiles:         make(map[int]mahjong.Tile),
		HuMultis:        make(map[int]int64),
		HuPlayers:       []int{},
		DecisionHistory: []Decision{},
		ActionHistory:   []ActionRecord{},
		CallData:        make(map[int32]*pbmj.CallData),
	}
}

func (s *GameState) RecordDecision(operate int, tile mahjong.Tile, obs []float32) {
	record := Decision{
		Operate: operate,
		Tile:    tile,
		Obs:     obs,
	}

	s.DecisionHistory = append(s.DecisionHistory, record)
}

// RecordExecutedAction 记录实现操作历史（根据 ack 记录实际执行的操作，用于生成特征，限制100条）
func (s *GameState) RecordAction(seat int, operate int, tile mahjong.Tile) {
	record := ActionRecord{
		Seat:      seat,
		Operate:   operate,
		TileIndex: mahjong.ToIndex(tile),
	}
	s.ActionHistory = append(s.ActionHistory, record)

	if len(s.ActionHistory) > HistorySteps {
		s.ActionHistory = s.ActionHistory[len(s.ActionHistory)-HistorySteps:]
	}
}

func (s *GameState) ToRichFeature() *RichFeature {
	r := &RichFeature{
		Hand:          [34]float32{},
		PlayerLacks:   [4][3]float32{},
		CurrentSeat:   [4]float32{},
		TotalTiles:    0.0,
		Operates:      [5]float32{},
		ActionHistory: [HistoryDim]float32{}, // 历史操作序列初始化为0
	}

	// 手牌
	for tile, count := range s.Hand {
		r.Hand[mahjong.ToIndex(tile)] = float32(count)
	}

	// 玩家缺门花色信息（one-hot编码）
	for seat := range 4 {
		lackColor := s.PlayerLacks[seat]
		if lackColor >= mahjong.ColorCharacter && lackColor <= mahjong.ColorDot {
			// 将花色转换为0-2的索引（万=0，条=1，筒=2）
			colorIndex := int(lackColor - mahjong.ColorCharacter)
			if colorIndex >= 0 && colorIndex < 3 {
				r.PlayerLacks[seat][colorIndex] = 1.0
			}
		}
	}

	// 当前玩家座位号（one-hot编码）
	if s.CurrentSeat >= 0 && s.CurrentSeat < 4 {
		r.CurrentSeat[s.CurrentSeat] = 1.0
	}

	// 总牌张数（归一化到0-1范围，假设最大牌数为144）
	if s.TotalTiles > 0 {
		r.TotalTiles = float32(s.TotalTiles) / 144.0
	}

	// 设置可执行操作（one-hot编码）
	if s.Operates != nil {
		if s.Operates.HasOperate(mahjong.OperateDiscard) {
			r.Operates[0] = 1.0
		}
		if s.Operates.HasOperate(mahjong.OperateHu) {
			r.Operates[1] = 1.0
		}
		if s.Operates.HasOperate(mahjong.OperatePon) {
			r.Operates[2] = 1.0
		}
		if s.Operates.HasOperate(mahjong.OperateKon) {
			r.Operates[3] = 1.0
		}
		if s.Operates.HasOperate(mahjong.OperatePass) {
			r.Operates[4] = 1.0
		}
	}

	// 编码历史操作序列（最多 HistorySteps 步，包含所有玩家的操作）
	historyLen := len(s.ActionHistory)
	startIdx := 0
	if historyLen > HistorySteps {
		startIdx = historyLen - HistorySteps // 只取最近 N 步
	}

	for i := range HistorySteps {
		historyIdx := startIdx + i
		offset := i * HistoryStepDim

		if historyIdx < historyLen && historyIdx >= 0 {
			rec := s.ActionHistory[historyIdx]

			// 编码操作类型（5维 one-hot）
			opIdx := 0
			switch rec.Operate {
			case mahjong.OperateDiscard:
				opIdx = 0
			case mahjong.OperateHu:
				opIdx = 1
			case mahjong.OperatePon:
				opIdx = 2
			case mahjong.OperateKon:
				opIdx = 3
			case mahjong.OperatePass:
				opIdx = 4
			default:
				opIdx = 4 // 默认 Pass
			}
			if opIdx >= 0 && opIdx < 5 {
				r.ActionHistory[offset+opIdx] = 1.0
			}

			// 编码玩家座位（4维 one-hot）
			if rec.Seat >= 0 && rec.Seat < 4 {
				r.ActionHistory[offset+5+rec.Seat] = 1.0
			}

			// 编码牌索引（34维 one-hot）
			if rec.TileIndex >= 0 && rec.TileIndex < 34 {
				r.ActionHistory[offset+5+4+rec.TileIndex] = 1.0
			}
		}
		// 如果历史不足，保持为0（已初始化）
	}

	return r
}
