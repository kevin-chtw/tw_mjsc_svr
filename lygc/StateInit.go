package lygc

import (
	"time"

	"github.com/kevin-chtw/tw_game_svr/mahjong"
)

type StateInit struct {
	*State
}

func NewStateInit(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateInit{
		State: NewState(game),
	}
}

func (s *StateInit) OnEnter() {
	s.State.GetMessager().SendDebugString("hello", -1)
	s.State.GetPlay().Initialize()
	s.State.GetMessager().SendGameStartAck()
	s.State.GetMessager().SendPlaceAck()

	s.AsyncTimer(time.Second, func() { s.GetGame().SetNextState(NewStateDeal) })
}
