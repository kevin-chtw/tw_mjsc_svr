package mahjong

import (
	"github.com/kevin-chtw/tw_game_svr/game"
)

type IGame interface {
	OnStart()
}

type Game struct {
	IGame
	*game.Table
	timer     *Timer
	CurState  IState
	nextState IState

	growSystem *GrowSystem
	config     *Config

	record  *Record
	play    *Play
	players []*Player
	nbSeat  ISeatID

	jsonData    interface{} // 用于存储临时数据
	requestIDs  []int       // 记录每个玩家的请求ID
	increasedID int         // 当前请求ID

}

func NewGame(subGame IGame) *Game {
	return &Game{
		IGame:       subGame,
		increasedID: 1,
	}
}

func (g *Game) OnGameBegin(table *game.Table) {
	g.Table = table
	g.players = make([]*Player, table.PlayerCount)
	for _, p := range table.Players {
		p.SetStatus(game.PlayerStatusPlaying)
		g.players[p.SeatNum] = NewPlayer(g, p)
	}

	g.requestIDs = make([]int, len(g.players))

	g.timer = NewTimer(g)        // 传入Game指针
	g.growSystem = &GrowSystem{} // 使用空结构体初始化
	g.config = NewConfig()
	g.record = NewRecord(g) // 传入Game指针
	g.play = NewPlay(g)
	g.nbSeat = g.initNewbieSeat()
	g.IGame.OnStart()
}

func (g *Game) OnCreateRecord() *Record {
	return NewRecord(g)
}

func (g *Game) OnCreatePlay() *Play {
	return NewPlay(g)
}

func (g *Game) GetPlay() *Play {
	return g.play
}

func (g *Game) OnGetDefaultProperties() []int {
	// 抽象方法，由子类实现
	return nil
}

func (g *Game) OnEnd(status EGameOverStatus) {
	// 游戏结束处理
}

func (g *Game) GetTimer() *Timer {
	return g.timer
}

func (g *Game) GetPlayer(seat ISeatID) *Player {
	if seat >= 0 && int(seat) < len(g.players) {
		return g.players[seat]
	}
	return nil
}

func (g *Game) GetGrowSystem() *GrowSystem {
	return g.growSystem
}

func (g *Game) GetConfig() *Config {
	return g.config
}

// GetLogger 获取游戏日志记录器
func (g *Game) GetLogger() Logger {
	// 返回默认实现的日志记录器
	return &defaultLogger{}
}

// defaultLogger 默认日志实现
type defaultLogger struct{}

func (l *defaultLogger) Printf(format string, args ...interface{}) {
	// 默认实现为空
}

func (l *defaultLogger) Println(args ...interface{}) {
	// 默认实现为空
}

func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	// 默认实现为空
}

func (g *Game) IsFDTable() bool {
	// 判断是否为房卡桌
	return false
}

func (g *Game) IsCoinMatch() bool {
	// 判断是否为金币场
	return false
}

func (g *Game) GetScoreBase() int {
	// 获取基础分
	return 1
}

func (g *Game) GetIniPlayerCount() int {
	// 获取初始玩家数量
	return len(g.players)
}

func (g *Game) FinishGame(status EGameOverStatus) {
	// 结束游戏
	g.OnEnd(status)
}

func (g *Game) IsValidSeat(seat ISeatID) bool {
	return seat >= 0 && int(seat) < len(g.players)
}

func (g *Game) initNewbieSeat() ISeatID {
	// 初始化新手座位
	return SeatNull
}

func (g *Game) SetNextState(newFn func(IGame, ...any) IState, args ...any) {
	g.nextState = CreateState(newFn, g.IGame, args...)
}
