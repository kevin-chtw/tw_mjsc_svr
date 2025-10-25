package mjsc

import (
	"time"

	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
)

type StateInit struct {
	*State
}

func NewStateInit(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateInit{
		State: NewState(game),
	}
}

func (s *StateInit) OnEnter() {
	s.game.play.Initialize(mahjong.NewPlayData)
	s.game.sender.SendGameStartAck()

	s.AsyncTimer(time.Second, func() { s.game.SetNextState(NewStateDeal) })
}
