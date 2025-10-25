package mjsc

import (
	"errors"

	"github.com/kevin-chtw/tw_common/gamebase/game"
	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/kevin-chtw/tw_common/utils"
	"github.com/kevin-chtw/tw_proto/game/pbmj"
	"github.com/kevin-chtw/tw_proto/game/pbsc"
)

type Game struct {
	*mahjong.Game
	play       *Play
	sender     *Sender
	scorelator mahjong.Scorelator
}

func NewGame(t *game.Table, id int32) game.IGame {
	g := &Game{}
	g.Game = mahjong.NewGame(g, t, id)
	g.play = NewPlay(g)
	g.sender = NewSender(g)
	g.scorelator = mahjong.NewScorelatorOnce(g.Game, mahjong.ScoreTypeMinScore)

	return g
}

func (g *Game) OnStart() {
	g.Game.SetNextState(NewStateInit)
}

func (g *Game) OnReqMsg(player *game.Player, data []byte) error {
	var msg pbsc.SCReq
	if err := utils.Unmarshal(player.Ctx, data, &msg); err != nil {
		return err
	}

	req, err := msg.Req.UnmarshalNew()
	if err != nil {
		return err
	}

	trust, ok := req.(*pbmj.MJTrustReq)
	if ok && !trust.GetTrust() {
		g.GetPlayer(player.GetSeat()).SetTrusted(false)
		return nil
	}

	if g.CurState == nil {
		return errors.New("invalid state")
	}
	return g.CurState.OnPlayerMsg(player.GetSeat(), req)
}
