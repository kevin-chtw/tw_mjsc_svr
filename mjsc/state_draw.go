package mjsc

import "github.com/kevin-chtw/tw_common/mahjong"

type StateDraw struct {
	*State
}

func NewStateDraw(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateDraw{
		State: NewState(game),
	}
}

func (s *StateDraw) OnEnter() {
	tile := s.game.play.Draw()
	if tile == mahjong.TileNull {
		s.game.SetNextState(NewStateLiuju)
		return
	}
	s.game.sender.SendDrawAck(tile)
	s.game.SetNextState(NewStateDiscard)
}
