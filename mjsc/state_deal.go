package mjsc

import (
	"github.com/kevin-chtw/tw_common/mahjong"
)

type StateDeal struct {
	*State
}

func NewStateDeal(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateDeal{
		State: NewState(game),
	}
}

func (s *StateDeal) OnEnter() {
	s.game.play.Deal()

	s.game.sender.SendOpenDoorAck()
	s.game.sender.SendAnimationAck()
	s.WaitAni(func() { s.game.SetNextState(NewStateDiscard) })
}
