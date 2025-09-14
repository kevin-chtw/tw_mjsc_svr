package mjsc

import (
	"errors"

	"github.com/kevin-chtw/tw_common/game"
	"github.com/kevin-chtw/tw_common/mahjong"
	"github.com/kevin-chtw/tw_proto/scproto"
	"google.golang.org/protobuf/encoding/protojson"
)

type Game struct {
	*mahjong.Game
	Play       *Play
	messager   *Messager
	scorelator *mahjong.Scorelator
}

func NewGame(t *game.Table) game.IGame {
	g := &Game{}
	g.Game = mahjong.NewGame(g, t)
	g.Play = NewPlay(g)
	g.messager = NewMessager(g)
	g.scorelator = mahjong.NewScroelator(g.Game)

	return g
}

func (g *Game) OnStart() {
	g.Game.SetNextState(NewStateInit)
}

func (g *Game) OnReqMsg(seat int32, data []byte) error {
	var msg scproto.SCReq
	if err := protojson.Unmarshal(data, &msg); err != nil {
		return err
	}

	if trust := msg.GetScTrustReq(); trust != nil && !trust.GetTrust() {
		g.GetPlayer(seat).SetTrusted(false)
		return nil
	}

	if g.CurState == nil {
		return errors.New("invalid state")
	}
	return g.CurState.OnPlayerMsg(seat, &msg)
}

func (g *Game) GetMessager() *Messager {
	return g.messager
}

func (g *Game) GetScorelator() *mahjong.Scorelator {
	return g.scorelator
}
