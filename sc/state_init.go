package sc

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
	s.GetPlay().Initialize()
	s.State.GetMessager().sendGameStartAck()

	s.AsyncTimer(time.Second, func() { s.game.SetNextState(NewStateDeal) })
}
