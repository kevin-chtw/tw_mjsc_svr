package mjsc

import "github.com/kevin-chtw/tw_common/mahjong"

type StateLiuju struct {
	*State
}

func NewStateLiuju(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateLiuju{
		State: NewState(game),
	}
}

func (s *StateLiuju) OnEnter() {
	s.onPlayerLiuJu()
	s.WaitAni(s.game.OnGameOver)
}

func (s *StateLiuju) onPlayerLiuJu() {
	// 实现流局处理逻辑
}
