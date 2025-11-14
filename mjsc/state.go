package mjsc

import (
	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/kevin-chtw/tw_mjsc_svr/ai"
)

type State struct {
	*mahjong.State
	game *Game
}

func NewState(game mahjong.IGame) *State {
	g := game.(*Game)
	return &State{
		State: mahjong.NewState(g.Game, g.sender.Sender),
		game:  g,
	}
}

// WaitAni 覆盖基类方法：训练模式下跳过动画等待
func (s *State) WaitAni(reqFn func()) {
	if ai.IsTrainingMode() {
		// 训练模式：立即执行，不等待动画
		reqFn()
		return
	}

	// 生产模式：等待5秒动画
	s.State.WaitAni(reqFn)
}
