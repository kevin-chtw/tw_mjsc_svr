package sc

import (
	"time"

	"github.com/kevin-chtw/tw_game_svr/mahjong"
)

type StateDeal struct {
	*State
}

var Deal = StateDeal{}

func NewStateDeal(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateDeal{
		State: NewState(game),
	}
}

func (s *StateDeal) OnEnter() {
	s.GetPlay().DoDeal()

	s.GetMessager().SendOpenDoorAck()
	s.GetMessager().SendBeginAnimalAck()
	s.AsyncTimer(time.Second*5, func() { s.game.Game.SetNextState(NewStateDiscard) })
}
