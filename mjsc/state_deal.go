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
	if s.game.GetRule().GetValue(RuleDiscardTime) != 0 {
		s.WaitAni(func() { s.game.SetNextState(NewStateSwapTiles) })
	} else {
		s.WaitAni(func() { s.game.SetNextState(NewStateDingque) })
	}
}
