package mjsc

import (
	"errors"

	"github.com/kevin-chtw/tw_common/game"
	"github.com/kevin-chtw/tw_common/mahjong"
	"github.com/kevin-chtw/tw_proto/scproto"
	"google.golang.org/protobuf/proto"
)

type Game struct {
	*mahjong.Game
	Play       *Play
	messager   *Messager
	scorelator mahjong.Scorelator
}

func NewGame(t *game.Table, id int32) game.IGame {
	g := &Game{}
	g.Game = mahjong.NewGame(g, t, id)
	g.Play = NewPlay(g)
	g.messager = NewMessager(g)
	g.scorelator = mahjong.NewScorelatorOnce(g.Game, mahjong.ScoreTypeMinScore)

	return g
}

func (g *Game) OnStart() {
	g.Game.SetNextState(NewStateInit)
}

func (g *Game) OnReqMsg(seat int32, data []byte) error {
	var msg scproto.SCReq
	if err := proto.Unmarshal(data, &msg); err != nil {
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

func (g *Game) GetScorelator() mahjong.Scorelator {
	if g.scorelator == nil {
		g.scorelator = mahjong.NewScorelatorOnce(g.Game, mahjong.ScoreTypeMinScore)
	}
	return g.scorelator
}
