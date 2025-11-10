package ai

import "github.com/kevin-chtw/tw_common/gamebase/mahjong"

// GameState 小型状态快照
type GameState struct {
	Operates    *mahjong.Operates       // 可执行操作
	CurrentSeat int                     // 当前玩家座位号（0-3）
	TotalTiles  int                     // 总牌张数
	SelfTurn    bool                    // 是否自己回合
	LastTile    mahjong.Tile            // 最近打出的牌
	Hand        map[mahjong.Tile]int    // 手牌（牌->数量）
	PonTiles    map[int][]mahjong.Tile  // 碰牌信息（玩家ID->碰的牌列表）
	GangTiles   map[int][]mahjong.Tile  // 杠牌信息（玩家ID->杠的牌列表）
	HuTiles     map[int]mahjong.Tile    // 胡牌信息（玩家ID->胡的牌）
	PlayerLacks [4]mahjong.EColor       // 四个玩家的缺门花色
	PlayerMelds [4]map[mahjong.Tile]int // 各玩家的副露（座位号->牌->数量）
	HuPlayers   []int                   // 胡牌玩家ID列表
}

func NewGameState() *GameState {
	return &GameState{
		TotalTiles:  136,
		Operates:    mahjong.NewOperates(int32(mahjong.OperateNone)),
		Hand:        make(map[mahjong.Tile]int),
		PonTiles:    make(map[int][]mahjong.Tile),
		GangTiles:   make(map[int][]mahjong.Tile),
		HuTiles:     make(map[int]mahjong.Tile),
		PlayerMelds: [4]map[mahjong.Tile]int{},
		HuPlayers:   []int{},
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
	for seat, tiles := range s.GangTiles {
		if seat >= 0 && seat < 4 {
			for _, tile := range tiles {
				r.GangInfo[seat][mahjong.ToIndex(tile)] = 1.0
			}
		}
	}

	// 胡牌信息
	for seat, tile := range s.HuTiles {
		if seat >= 0 && seat < 4 {
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

	return r
}
