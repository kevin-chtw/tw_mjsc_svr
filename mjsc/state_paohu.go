package mjsc

import (
	"time"

	"github.com/kevin-chtw/tw_common/mahjong"
)

type StatePaohu struct {
	*StateResult
}

func NewStatePaohu(game mahjong.IGame, args ...any) mahjong.IState {
	s := &StatePaohu{
		StateResult: NewStateResult(game),
	}
	s.huSeats = args[0].([]int32)
	return s
}

func (s *StatePaohu) OnEnter() {
	multiples := s.GetPlay().Zimo()
	s.game.GetMessager().sendHuAck(s.huSeats, s.GetPlay().GetCurSeat())
	s.game.GetScorelator().AddMultiple(mahjong.ScoreReasonHu, multiples)
	s.game.GetScorelator().Calculate()
	s.game.GetMessager().sendResult(true, 0, 0)

	s.game.GetMessager().sendAnimationAck()
	s.AsyncMsgTimer(s.onMsg, time.Second*5, s.game.OnGameOver)
}
