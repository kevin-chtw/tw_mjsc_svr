package ai

import (
	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/kevin-chtw/tw_proto/game/pbmj"
)

// DecisionRecord 决策记录（使用 Decision 作为别名，保持向后兼容）
type DecisionRecord = Decision

// ActionRecord 实现操作记录（根据 ack 记录的实际操作，用于生成特征）
type ActionRecord struct {
	Seat      int // 执行操作的玩家座位号（0-3）
	Operate   int // 操作类型
	TileIndex int // 牌索引
}

// GameState 小型状态快照
type GameState struct {
	Operates        *mahjong.Operates       // 可执行操作
	CurrentSeat     int                     // 当前玩家座位号（0-3）
	TotalTiles      int                     // 总牌张数
	SelfTurn        bool                    // 是否自己回合
	LastTile        mahjong.Tile            // 最近打出的牌
	Hand            map[mahjong.Tile]int    // 手牌（牌->数量）
	PonTiles        map[int][]mahjong.Tile  // 碰牌信息（玩家ID->碰的牌列表）
	KonTiles        map[int][]mahjong.Tile  // 杠牌信息（玩家ID->杠的牌列表）
	HuTiles         map[int]mahjong.Tile    // 胡牌信息（玩家ID->胡的牌）
	HuMultis        map[int]int64           // 胡牌番数（玩家ID->番数）
	PlayerLacks     [4]mahjong.EColor       // 四个玩家的缺门花色
	PlayerMelds     [4]map[mahjong.Tile]int // 各玩家的副露（座位号->牌->数量）
	HuPlayers       []int                   // 胡牌玩家ID列表
	DecisionHistory []DecisionRecord        // 决策历史记录（用于训练，不限制长度，不生成特征）
	ActionHistory   []ActionRecord          // 实现操作历史记录（用于生成特征，限制60条）
	CallData        map[int32]*pbmj.CallData
	// 终局统计信息
	FinalScore float32 // 最终得分（包含点炮惩罚）
	IsLiuJu    bool    // 是否流局
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
		PlayerMelds:     [4]map[mahjong.Tile]int{make(map[mahjong.Tile]int), make(map[mahjong.Tile]int), make(map[mahjong.Tile]int), make(map[mahjong.Tile]int)},
		HuPlayers:       []int{},
		DecisionHistory: []DecisionRecord{},
		ActionHistory:   []ActionRecord{},
		CallData:        make(map[int32]*pbmj.CallData),
	}
}

// RecordDecision 记录决策历史（AI做决策时调用，保存 Obs 用于训练，不限制长度，不生成特征）
func (s *GameState) RecordDecision(operate int, tile mahjong.Tile) {
	record := DecisionRecord{
		Operate: operate,
		Tile:    tile,
		Obs:     s.ToRichFeature().ToVector(), // 保存状态特征用于训练
	}
	s.DecisionHistory = append(s.DecisionHistory, record)
	// 不限制长度，保留所有决策记录用于训练
}

// RecordExecutedAction 记录实现操作历史（根据 ack 记录实际执行的操作，用于生成特征，限制100条）
func (s *GameState) RecordAction(seat int, operate int, tile mahjong.Tile) {
	record := ActionRecord{
		Seat:      seat,
		Operate:   operate,
		TileIndex: mahjong.ToIndex(tile),
	}
	s.ActionHistory = append(s.ActionHistory, record)

	// 限制历史记录最大长度为 60，超过时只保留最后60条（减少内存和计算）
	if len(s.ActionHistory) > 60 {
		s.ActionHistory = s.ActionHistory[len(s.ActionHistory)-60:]
	}
}

func (s *GameState) ToRichFeature() *RichFeature {
	r := &RichFeature{
		Hand:          [34]float32{},
		Furo:          [4][34]float32{}, // 使用Furo字段存储各玩家副露
		PonInfo:       [4][34]float32{},
		GangInfo:      [4][34]float32{},
		HuInfo:        [4][34]float32{},
		PlayerActions: [4]float32{},
		PlayerLacks:   [4][3]float32{},
		CurrentSeat:   [4]float32{},
		TotalTiles:    0.0,
		SelfTurn:      0.0,
		Operates:      [5]float32{},
		ActionHistory: [HistoryDim]float32{}, // 历史操作序列初始化为0
	}

	// 手牌
	for tile, count := range s.Hand {
		r.Hand[mahjong.ToIndex(tile)] = float32(count)
	}

	// 各玩家副露信息（存储到Furo字段）
	for seat := range 4 {
		if s.PlayerMelds[seat] != nil {
			for tile, count := range s.PlayerMelds[seat] {
				r.Furo[seat][mahjong.ToIndex(tile)] = float32(count)
			}
		}
	}

	// 碰牌信息
	for seat, tiles := range s.PonTiles {
		if seat >= 0 && seat < 4 {
			for _, tile := range tiles {
				r.PonInfo[seat][mahjong.ToIndex(tile)] = 1.0
			}
		}
	}

	// 杠牌信息
	for seat, tiles := range s.KonTiles {
		if seat >= 0 && seat < 4 {
			for _, tile := range tiles {
				r.GangInfo[seat][mahjong.ToIndex(tile)] = 1.0
			}
		}
	}

	// 胡牌信息
	for seat, tile := range s.HuTiles {
		if tile > 0 && seat >= 0 && seat < 4 {
			r.HuInfo[seat][mahjong.ToIndex(tile)] = 1.0
		}
	}

	// 玩家动作状态（胡牌玩家标记为1）
	for _, seat := range s.HuPlayers {
		if seat >= 0 && seat < 4 {
			r.PlayerActions[seat] = 1.0
		}
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

	// 是否自己回合（0或1）
	if s.SelfTurn {
		r.SelfTurn = 1.0
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

	for i := 0; i < HistorySteps; i++ {
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
