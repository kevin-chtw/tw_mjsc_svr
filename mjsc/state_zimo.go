package mjsc

import (
	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
)

type StateZimo struct {
	*State
	huSeats []int32
}

func NewStateZimo(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateZimo{
		State:   NewState(game),
		huSeats: make([]int32, 0),
	}
}

func (s *StateZimo) OnEnter() {
	s.huSeats = append(s.huSeats, s.game.play.GetCurSeat())
	multiples := s.game.play.PaoHu(s.huSeats)
	s.game.sender.SendHuAck(s.huSeats, mahjong.SeatNull)
	scores := s.game.scorelator.Calculate(multiples)
	s.game.sender.SendScoreChangeAck(mahjong.ScoreReasonHu, scores, s.game.play.GetCurTile(), mahjong.SeatNull, s.huSeats)
	s.game.sender.SendResult(false)

	s.WaitAni(s.game.OnGameOver)
}
