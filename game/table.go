package game

import (
	"context"
	"errors"
	"time"

	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/kevin-chtw/tw_proto/sproto"
	"github.com/sirupsen/logrus"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
)

const (
	// TableStatusPreparing 桌子状态：准备中
	TableStatusPreparing = "preparing"
	// TableStatusPlaying 桌子状态：游戏中
	TableStatusPlaying = "playing"
	// TableStatusFinished 桌子状态：已结束
	TableStatusFinished = "finished"
)

// Table 表示一个游戏桌实例
type Table struct {
	GameID        int32              // 游戏ID
	MatchID       int32              // 比赛ID
	TableID       int32              // 桌号
	MatchServerId string             // 匹配服务ID
	Players       map[string]*Player // 玩家ID -> Player
	Status        string             // "preparing", "playing", "finished"
	PlayerCount   int32              // 玩家数量
	CreateTime    int64
	timer         *time.Timer
	App           pitaya.Pitaya
}

// NewTable 创建新的游戏桌实例
func NewTable(gameId, matchID, tableID int32, app pitaya.Pitaya) *Table {
	return &Table{
		GameID:        gameId,
		MatchID:       matchID,
		TableID:       tableID,
		MatchServerId: "",
		Players:       make(map[string]*Player),
		PlayerCount:   0,
		Status:        TableStatusPreparing,
		CreateTime:    time.Now().Unix(),
		App:           app,
	}
}

// OnPlayerMsg 处理玩家消息
func (t *Table) OnPlayerMsg(ctx context.Context, player *Player, req *cproto.GameReq) {
	if req.GetEnterGameReq() != nil {
		t.handleEnterGame(ctx, player, req.GetEnterGameReq())
	}
	if req.GetTableMsgReq() != nil {
		t.handleTableMsg(ctx, player, req.GetTableMsgReq())
	}
}

// handleEnterGame 处理玩家进入游戏请求
func (t *Table) handleEnterGame(ctx context.Context, player *Player, req *cproto.EnterGameReq) {
	if !t.isOnTable(player.ID) {
		return // 玩家不在桌上
	}

	// 添加玩家到桌中
	t.Players[player.ID] = player
	t.broadcastTablePlayer(player)
	t.SendTablePlayer(player)
	player.SetStatus(PlayerStatusReady)

	// 检查是否满足开赛条件
	if t.isAllPlayersReady() {
		t.Status = TableStatusPlaying
		t.OnGameBegin()
	}
}

func (t *Table) broadcastTablePlayer(player *Player) {
	msg := t.NewMsg()
	msg.Ack = &cproto.GameAck_TablePlayerAck{
		TablePlayerAck: &cproto.TablePlayerAck{
			Playerid: player.ID,
			Seatnum:  player.SeatNum,
		},
	}
	t.Broadcast(msg)
	logrus.Infof("Player %s added to table %s", player.ID, t.TableID)
}

func (t *Table) SendTablePlayer(player *Player) {
	msg := t.NewMsg()
	for _, p := range t.Players {
		if p.ID != player.ID {
			msg.Ack = &cproto.GameAck_TablePlayerAck{
				TablePlayerAck: &cproto.TablePlayerAck{
					Playerid: p.ID,
					Seatnum:  p.SeatNum,
				},
			}
			t.SendMsg(msg, player.ID)
		}
	}
}

func (t *Table) NewMsg() *cproto.GameAck {
	return &cproto.GameAck{
		Serverid: t.App.GetServerID(),
		Gameid:   t.GameID,
		Tableid:  t.TableID,
		Matchid:  t.MatchID,
	}
}

func (t *Table) isOnTable(playerID string) bool {
	if _, ok := t.Players[playerID]; ok {
		return true
	}
	return false
}

func (t *Table) isAllPlayersReady() bool {
	for _, player := range t.Players {
		if player.Status != PlayerStatusReady {
			return false
		}
	}
	return true
}

func (t *Table) handleTableMsg(ctx context.Context, player *Player, req *cproto.TableMsgReq) {

}

// HandleStartGame 处理开始游戏请求
func (t *Table) HandleStartGame(ctx context.Context, req *sproto.AddTableReq) {
	t.Status = TableStatusPlaying
	t.PlayerCount = req.PlayerCount

	// 设置10秒后自动结束游戏
	t.timer = time.AfterFunc(10*time.Second, func() {
		t.Status = "finished"
		if t.timer != nil {
			t.timer.Stop()
			t.timer = nil
		}
	})
}

func (t *Table) HandleAddPlayer(ctx context.Context, req *sproto.AddPlayerReq) error {
	if t.isOnTable(req.Playerid) {
		return errors.New("Player already on table")
	}

	player := playerManager.GetPlayer(req.Playerid)
	player.SetSeat(req.Seatnum)
	player.AddScore(req.Score)
	t.Players[req.Playerid] = player
	return nil
}

func (t *Table) HandleCancelTable(ctx context.Context, req *sproto.CancelTableReq) error {
	if t.Status != TableStatusPreparing {
		logrus.Warnf("Table %d is not in preparing status, cannot cancel", t.TableID)
		return errors.New("table not in preparing status")
	}
	t.Status = TableStatusFinished
	logrus.Infof("Table %d cancelled", t.TableID)
	// 通知匹配服务
	if err := t.SendToMatch(); err != nil {
		logrus.Errorf("Failed to send cancel request to match service: %v", err)
		return errors.New("failed to send cancel request to match service")
	}
	// 清理玩家状态
	for _, player := range t.Players {
		player.SetStatus(PlayerStatusOffline)
	}
	t.Players = make(map[string]*Player) // 清空玩家列表
	for _, player := range t.Players {
		playerManager.Delete(player.ID) // 从玩家管理器中删除玩家
	}
	tableManager.Delete(t.MatchID, t.TableID) // 从桌子管理器中删除
	return nil
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

func (t *Table) Broadcast(msg *cproto.GameAck) {
	players := make([]string, 0, len(t.Players))
	for _, player := range t.Players {
		if player.Status != PlayerStatusOffline && player.Status != PlayerStatusUnEnter {
			players = append(players, player.ID)
		}
	}
	if m, err := t.App.SendPushToUsers("gamemsg", msg, players, "proxy"); err != nil {
		logrus.Errorf("send game message failed: %v", err)
	} else {
		logrus.Infof("broadcast game message to players: %v", m)
	}
}

func (t *Table) SendMsg(msg *cproto.GameAck, playerID string) error {
	if m, err := t.App.SendPushToUsers("gamemsg", msg, []string{playerID}, "proxy"); err != nil {
		logrus.Errorf("send game message to player %s failed: %v", playerID, err)
		return err
	} else {
		logrus.Infof("send game message to player %s: %v", playerID, m)
	}
	return nil
}

func (t *Table) OnGameBegin() {
	// 这里可以添加游戏开始的逻辑，比如发牌、初始化游戏状态等
	for _, player := range t.Players {
		player.SetStatus("playing")
	}
}
