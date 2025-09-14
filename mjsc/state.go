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
		State: mahjong.NewState(g.Game),
		game:  g,
	}
}
func (s *State) GetGame() *Game {
	return s.game
}

func (s *State) GetPlay() *Play {
	return s.game.Play
}

func (s *State) GetMessager() *Messager {
	return s.game.GetMessager()
}
