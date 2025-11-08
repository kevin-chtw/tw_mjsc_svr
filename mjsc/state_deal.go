package mjsc

import (
	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
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
	if s.game.GetRule().GetValue(RuleSwapTile) != 0 {
		s.WaitAni(func() { s.game.SetNextState(NewStateSwapTiles) })
	} else if s.game.GetRule().GetValue(RuleLiangMen) != 0 {
		for seat := range s.game.GetPlayerCount() {
			s.game.play.FreshCallData(seat)
			s.game.sender.SendCallDataAck(seat)
		}
		s.game.SetNextState(NewStateDiscard)
	} else {
		s.WaitAni(func() { s.game.SetNextState(NewStateDingque) })
	}
}
