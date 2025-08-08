package sc

import (
	"github.com/kevin-chtw/tw_game_svr/mahjong"
)

type State struct {
	*mahjong.State
	game          *Game
	requestStatus []bool // 玩家请求状态位图
}

func NewState(game mahjong.IGame) *State {
	g := game.(*Game)
	return &State{
		State:         mahjong.NewState(g.Game),
		requestStatus: make([]bool, g.Game.GetPlayerCount()),
		game:          g,
	}
}
func (s *State) GetGame() *Game {
	return s.game
}

func (s *State) GetPlay() *Play {
	return s.game.Play.(*Play)
}

func (s *State) GetMessager() *Messager {
	return s.game.GetMessager()
}

func (s *State) SetRequestStatus(seat int, status bool) {
	if seat >= 0 && seat < len(s.requestStatus) {
		s.requestStatus[seat] = status
	}
}

func (s *State) GetRequestStatus(seat int) bool {
	if seat >= 0 && seat < len(s.requestStatus) {
		return s.requestStatus[seat]
	}
	return false
}
