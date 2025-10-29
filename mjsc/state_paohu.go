package mjsc

import (
	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
)

type StatePaohu struct {
	*State
	huSeats []int32
}

func NewStatePaohu(game mahjong.IGame, args ...any) mahjong.IState {
	s := &StatePaohu{
		State: NewState(game),
	}
	s.huSeats = args[0].([]int32)
	return s
}

func (s *StatePaohu) OnEnter() {
	multiples := s.game.play.PaoHu(s.huSeats)
	scores := s.game.scorelator.Calculate(multiples)
	s.game.sender.SendScoreChangeAck(mahjong.ScoreReasonHu, scores, s.game.play.GetCurTile(), s.game.play.GetCurSeat(), s.huSeats)
	s.game.sender.SendResult(false)

	s.WaitAni(s.game.OnGameOver)
}
