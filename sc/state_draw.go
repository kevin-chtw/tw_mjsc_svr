package sc

import "github.com/kevin-chtw/tw_game_svr/mahjong"

type StateDraw struct {
	*State
}

func NewStateDraw(game mahjong.IGame) *StateDraw {
	return &StateDraw{
		State: NewState(game),
	}
}

func (s *StateDraw) OnEnter() {
	// 实现抽牌状态进入逻辑
}
