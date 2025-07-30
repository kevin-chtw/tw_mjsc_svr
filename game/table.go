package game

import (
	"context"
	"time"

	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/kevin-chtw/tw_proto/sproto"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
)

// Table 表示一个游戏桌实例
type Table struct {
	TableID       string
	MatchID       string
	MatchServerId string // 匹配服务ID
	Players       []string
	Status        string // "preparing", "playing", "finished"
	CreateTime    int64
	timer         *time.Timer
	App           pitaya.Pitaya
}

// NewTable 创建新的游戏桌实例
func NewTable(matchID, tableID string, app pitaya.Pitaya) *Table {
	return &Table{
		TableID:       tableID,
		MatchID:       matchID,
		MatchServerId: "",
		Players:       make([]string, 0),
		Status:        "preparing",
		CreateTime:    time.Now().Unix(),
		App:           app,
	}
}

// OnPlayerMsg 处理玩家消息
func (t *Table) OnPlayerMsg(ctx context.Context, playerID string, data []byte) {
	// TODO: 实现具体的消息处理逻辑
}

// HandleStartGame 处理开始游戏请求
func (t *Table) HandleStartGame(ctx context.Context, req *sproto.StartGameReq) {
	t.Status = "playing"
	t.MatchServerId = req.MatchServerId
	// 设置10秒后自动结束游戏
	t.timer = time.AfterFunc(10*time.Second, func() {
		t.Status = "finished"
		if t.timer != nil {
			t.timer.Stop()
			t.timer = nil
		}
	})
}

// HandleCancelGame 处理取消游戏请求
func (t *Table) HandleCancelGame(ctx context.Context, req *sproto.CancelMatchReq) {
	t.Status = "finished"
	// 取消定时器
	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
}

func (t *Table) SendToMatch() error {
	rsp := &cproto.CommonResponse{Err: cproto.ErrCode_OK}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := t.App.RPCTo(ctx, t.MatchServerId, "match.game.message", rsp, &sproto.Match2GameAck{}); err != nil {
		t.Status = "preparing" // 回滚状态
		return err
	}
	return nil
}
