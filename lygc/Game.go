package lygc

import (
	"github.com/kevin-chtw/tw_game_svr/game"
	"github.com/kevin-chtw/tw_game_svr/mahjong"
	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/kevin-chtw/tw_proto/lygcproto"
	"google.golang.org/protobuf/encoding/protojson"
)

type Game struct {
	*mahjong.Game
	messager  *Messager
	checker   *mahjong.CheckerOnce
	roundData *LastHandData
}

func NewGame() *Game {
	g := &Game{
		roundData: &LastHandData{},
	}
	g.Game = mahjong.NewGame(g)
	g.messager = NewMessager(g)
	g.checker = mahjong.NewCheckerOnce(g.Game, mahjong.ScoreTypeMinScore)
	return g
}

func (g *Game) OnStart() {
	g.Game.SetNextState(NewStateInit)
}

func (g *Game) OnPlayerMsg(player *game.Player, req *cproto.TableMsgReq) {
	if req == nil {
		return
	}
	seat := player.SeatNum
	if seat < 0 || seat >= g.Game.Table.PlayerCount {
		return
	}

	if len(req.Msg) == 0 {
		return
	}

	var msg lygcproto.LYGCReq
	if err := protojson.Unmarshal(req.Msg, &msg); err != nil {
		return
	}

	if trust := msg.GetLygcTrustReq(); trust != nil && !trust.GetTrust() {
		// TODO 取消托管
		return
	}

	if g.Game.CurState != nil {
		g.Game.CurState.OnPlayerMsg(player, &msg)
	}
}

func (g *Game) GetMessager() *Messager {
	return g.messager
}

func (g *Game) GetChecker() *mahjong.CheckerOnce {
	return g.checker
}

func (g *Game) GetPlay() *Play {
	return nil
}

func (g *Game) SendMsg(ack *cproto.GameAck, seat int) {
	// 实现发送消息逻辑
}

func (g *Game) GetRoundData() *LastHandData {
	return g.roundData
}

func (g *Game) UpdateRoundData() {
	// 实现更新回合数据逻辑
}
