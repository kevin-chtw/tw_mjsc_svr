package mjsc

import (
	"github.com/kevin-chtw/tw_common/mahjong"
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
