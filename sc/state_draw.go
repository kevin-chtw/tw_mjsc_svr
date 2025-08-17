package sc

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
	tile := s.GetPlay().Draw()
	if tile == mahjong.TileNull {
		s.game.SetNextState(NewStateLiuju)
		return
	}
	s.GetMessager().sendDrawAck(tile)
	s.game.SetNextState(NewStateDiscard)
}
