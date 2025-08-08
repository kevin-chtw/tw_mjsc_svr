package mahjong

import (
	"github.com/kevin-chtw/tw_game_svr/game"
	"github.com/kevin-chtw/tw_proto/cproto"
)

type IGame interface {
	CreatePlay() IPlay
	OnStart()
	OnReqMsg(seat int32, req *cproto.TableMsgReq)
}

type Game struct {
	IGame
	*game.Table
	timer       *Timer
	CurState    IState
	nextState   IState
	config      *Config
	players     []*Player
	increasedID int   // 当前请求ID
	requestIDs  []int // 记录每个玩家的请求ID

	Play IPlay
}

func NewGame(subGame IGame, t *game.Table) *Game {
	g := &Game{
		IGame:       subGame,
		Table:       t,
		timer:       NewTimer(),
		players:     make([]*Player, t.GetPlayerCount()),
		increasedID: 1,
		requestIDs:  make([]int, t.GetPlayerCount()),
	}

	for i := int32(0); i < t.GetPlayerCount(); i++ {
		g.players[i] = NewPlayer(g, t.GetGamePlayer(i))
	}
	g.Play = subGame.CreatePlay()
	return g
}

func (g *Game) OnGameBegin() {
	g.Play.Initialize()
	g.IGame.OnStart()
	g.enterNextState()
}

func (g *Game) OnPlayerMsg(player *game.Player, req *cproto.TableMsgReq) {
	if req == nil {
		return
	}
	seat := player.Seat
	if !g.IsValidSeat(seat) {
		return
	}

	if len(req.Msg) == 0 {
		return
	}
	g.IGame.OnReqMsg(seat, req)
	g.enterNextState()
}

func (g *Game) GetTimer() *Timer {
	return g.timer
}

func (g *Game) GetPlayer(seat int32) *Player {
	if seat >= 0 && int(seat) < len(g.players) {
		return g.players[seat]
	}
	return nil
}

func (g *Game) SetNextState(newFn func(IGame, ...any) IState, args ...any) {
	g.nextState = newFn(g.IGame, args...)
}

func (g *Game) enterNextState() {
	for g.nextState != nil {
		g.CurState = g.nextState
		g.nextState = nil
		g.timer.Cancel()
		g.CurState.OnEnter()
	}
}
