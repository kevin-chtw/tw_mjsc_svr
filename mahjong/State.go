package mahjong

import (
	"time"

	"github.com/kevin-chtw/tw_game_svr/game"
)

type StateOpts struct {
	// 自定义选项
}

type IState interface {
	OnEnter()
	OnPlayerMsg(player *game.Player, message interface{})
}

func CreateState[T IState](newFn func(IGame, ...any) T, g IGame, args ...any) T {
	return newFn(g, args)
}

// State 麻将游戏状态基类
type State struct {
	game       *Game
	msgHandler func(seat ISeatID, req interface{})
}

// NewState 创建新的游戏状态
func NewState(game *Game) *State {
	return &State{
		game:       game,
		msgHandler: nil,
	}
}

// AsyncMsgTimer 设置异步消息定时器
func (s *State) AsyncMsgTimer(
	handler func(seat ISeatID, req interface{}),
	timeout time.Duration,
	onTimeout func(),
) {
	s.msgHandler = handler
	s.game.GetTimer().Schedule(timeout, onTimeout)
}

// AsyncTimer 设置异步定时器
func (s *State) AsyncTimer(timeout time.Duration, onTimeout func()) {
	s.msgHandler = nil
	s.game.GetTimer().Schedule(timeout, onTimeout)
}

// HandlePlayerMsg 处理玩家消息
func (s *State) OnPlayerMsg(player *game.Player, message interface{}) {
	if s.msgHandler != nil {
		s.msgHandler(ISeatID(player.SeatNum), message)
	}
}
