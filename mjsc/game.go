package mjsc

import (
	"errors"

	"github.com/kevin-chtw/tw_common/game"
	"github.com/kevin-chtw/tw_common/mahjong"
	"github.com/kevin-chtw/tw_proto/pbmj"
	"github.com/kevin-chtw/tw_proto/pbsc"
	"google.golang.org/protobuf/proto"
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

func (g *Game) OnReqMsg(seat int32, data []byte) error {
	var msg pbsc.SCReq
	if err := proto.Unmarshal(data, &msg); err != nil {
		return err
	}

	req, err := msg.Req.UnmarshalNew()
	if err != nil {
		return err
	}

	trust, ok := req.(*pbmj.MJTrustReq)
	if ok && !trust.GetTrust() {
		g.GetPlayer(seat).SetTrusted(false)
		return nil
	}

	if g.CurState == nil {
		return errors.New("invalid state")
	}
	return g.CurState.OnPlayerMsg(seat, req)
}
