package mjsc

import (
	"github.com/kevin-chtw/tw_common/game"
	"github.com/kevin-chtw/tw_common/mahjong"
	"github.com/kevin-chtw/tw_proto/scproto"
	"google.golang.org/protobuf/encoding/protojson"
)

type Game struct {
	*mahjong.Game
	messager   *Messager
	scorelator *mahjong.Scorelator
}

func NewGame(t *game.Table) game.IGame {
	g := &Game{}
	g.messager = NewMessager(g)
	g.scorelator = mahjong.NewScroelator(g.Game)
	g.Game = mahjong.NewGame(g, t)
	return g
}

func (g *Game) CreatePlay() mahjong.IPlay {
	return NewPlay(g)
}

func (g *Game) OnStart() {
	g.Game.SetNextState(NewStateInit)
}

func (g *Game) OnReqMsg(seat int32, data []byte) {
	var msg scproto.SCReq
	if err := protojson.Unmarshal(data, &msg); err != nil {
		return
	}

	if trust := msg.GetScTrustReq(); trust != nil && !trust.GetTrust() {
		g.GetPlayer(seat).SetTrusted(false)
		return
	}

	if g.Game.CurState != nil {
		g.Game.CurState.OnPlayerMsg(seat, &msg)
	}
}

func (g *Game) GetMessager() *Messager {
	return g.messager
}

func (g *Game) GetScorelator() *mahjong.Scorelator {
	return g.scorelator
}
