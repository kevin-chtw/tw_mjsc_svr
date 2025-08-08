package sc

import "github.com/kevin-chtw/tw_game_svr/mahjong"

type Play struct {
	*mahjong.Play
}

func NewPlay(game *Game) *Play {
	return &Play{
		Play: mahjong.NewPlay(game.Game),
	}
}
