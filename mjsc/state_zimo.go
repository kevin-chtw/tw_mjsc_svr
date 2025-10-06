package mjsc

import (
	"time"

	"github.com/kevin-chtw/tw_common/mahjong"
)

type StateZimo struct {
	*StateResult
}

func NewStateZimo(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateZimo{
		StateResult: NewStateResult(game),
	}
}

func (s *StateZimo) OnEnter() {
	s.huSeats = append(s.huSeats, s.game.play.GetCurSeat())
	multiples := s.game.play.PaoHu(s.huSeats)
	s.game.sender.SendHuAck(s.huSeats, mahjong.SeatNull)
	s.game.scorelator.AddMultiple(mahjong.ScoreReasonHu, multiples)
	s.game.scorelator.Calculate()
	s.game.sender.SendResult(false)

	s.game.sender.SendAnimationAck()
	s.AsyncMsgTimer(s.onMsg, time.Second*5, s.game.OnGameOver)
}
