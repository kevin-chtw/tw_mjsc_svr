package game

import (
	"context"
	"sync"

	"github.com/kevin-chtw/tw_proto/cproto"
)

// Player 表示游戏中的玩家
type Player struct {
	ID      string // 玩家唯一ID
	SeatNum int    // 座位号
	Score   int32  // 玩家积分
	Status  string // 玩家状态: "ready", "playing", "offline"
	mu      sync.RWMutex
}

// NewPlayer 创建新玩家实例
func NewPlayer(id string) *Player {
	return &Player{
		ID:     id,
		Status: "ready",
	}
}

// SetSeat 设置玩家座位号
func (p *Player) SetSeat(seatNum int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.SeatNum = seatNum
}

// AddScore 增加玩家积分
func (p *Player) AddScore(delta int32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Score += delta
}

// SetStatus 设置玩家状态
func (p *Player) SetStatus(status string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Status = status
}

// GetInfo 获取玩家信息
func (p *Player) GetInfo() (string, int, int32, string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.ID, p.SeatNum, p.Score, p.Status
}

// HandleMessage 处理玩家消息
func (p *Player) HandleMessage(ctx context.Context, req *cproto.GameReq) {
	table := tableManager.Get(req.Matchid, req.Tableid)
	if nil == table {
		return // 桌子不存在
	}
	table.OnPlayerMsg(ctx, p.ID, req.Data)
}
