package sc

import (
	"github.com/kevin-chtw/tw_game_svr/mahjong"
)

type StateDiscard struct {
	*State
	operates *mahjong.Operates
}

func NewStateDiscard(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateDiscard{
		State: NewState(game),
	}
}

func (s *StateDiscard) OnEnter() {
	s.operates = s.GetPlay().FetchSelfOperates()
}

func (s *StateDiscard) AutoOperate(isTimeout bool) {
	// 实现自动操作逻辑
}

func (s *StateDiscard) Discard(tile int) {
	// 实现弃牌逻辑
}

func (s *StateDiscard) Kon(tile int) {
	// 实现杠牌逻辑
}

func (s *StateDiscard) OnTimeout() {
	// 实现超时处理逻辑
}
