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
	s.huSeats = append(s.huSeats, s.GetPlay().GetCurSeat())
	multiples := s.GetPlay().PaoHu(s.huSeats)
	s.game.GetMessager().sendHuAck(s.huSeats, mahjong.SeatNull)
	s.game.GetScorelator().AddMultiple(mahjong.ScoreReasonHu, multiples)
	s.game.GetScorelator().Calculate()
	s.game.GetMessager().sendResult(false, mahjong.SeatNull, mahjong.SeatNull)

	s.game.GetMessager().sendAnimationAck()
	s.AsyncMsgTimer(s.onMsg, time.Second*5, s.game.OnGameOver)
}
