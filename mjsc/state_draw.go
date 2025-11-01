package mjsc

import "github.com/kevin-chtw/tw_common/gamebase/mahjong"

type StateDraw struct {
	*State
}

func NewStateDraw(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateDraw{
		State: NewState(game),
	}
}

func (s *StateDraw) OnEnter() {
	if s.game.GetRestCount() == 1 {
		s.game.sender.SendResult(false)
		s.WaitAni(s.game.OnGameOver)
		return
	}

	tile := s.game.play.Draw()
	if tile == mahjong.TileNull {
		s.game.sender.SendResult(true)
		s.WaitAni(s.game.OnGameOver)
		return
	}
	s.game.sender.SendDrawAck(tile)
	s.game.SetNextState(NewStateDiscard)
}
