package mjsc

import "github.com/kevin-chtw/tw_common/mahjong"

type StateLiuju struct {
	*StateResult
}

func NewStateLiuju(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateLiuju{
		StateResult: NewStateResult(game),
	}
}

func (s *StateLiuju) OnEnter() {
	s.onPlayerLiuJu()
}

func (s *StateLiuju) onPlayerLiuJu() {
	// 实现流局处理逻辑
}
