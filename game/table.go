package game

import (
	"context"
	"time"

	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/kevin-chtw/tw_proto/sproto"
	"github.com/sirupsen/logrus"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
)

type IGame interface {
	// OnGameBegin 游戏开始
	OnGameBegin(Table *Table)
	// OnPlayerMsg 处理玩家消息
	OnPlayerMsg(player *Player, req *cproto.TableMsgReq)
}

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
	app           pitaya.Pitaya
	MatchType     int32  // 0: 普通匹配, 1: 房卡模式
	ScoreBase     int32  // 分数基数
	GameCount     int32  // 游戏局数
	PlayerCount   int32  // 玩家数量
	GameConfig    string // 游戏配置
	game          IGame  // 游戏逻辑处理接口
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
		app:           app,
		game:          CreateGame(gameId),
	}
}

// OnPlayerMsg 处理玩家消息
func (t *Table) OnPlayerMsg(ctx context.Context, player *Player, req *cproto.GameReq) {
	if req.GetEnterGameReq() != nil {
		t.handleEnterGame(player, req.GetEnterGameReq())
	}
	if req.GetTableMsgReq() != nil {
		t.game.OnPlayerMsg(player, req.GetTableMsgReq())
	}
}

// handleEnterGame 处理玩家进入游戏请求
func (t *Table) handleEnterGame(player *Player, _ *cproto.EnterGameReq) {
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
		t.game.OnGameBegin(t)
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
		Serverid: t.app.GetServerID(),
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

// HandleStartGame 处理开始游戏请求
func (t *Table) HandleAddTable(ctx context.Context, req *sproto.AddTableReq) *sproto.AddTableAck {
	ack := &sproto.AddTableAck{
		ErrorCode: int32(0), // 成功
	}

	t.Status = TableStatusPreparing
	t.MatchType = req.GetMatchType()
	t.ScoreBase = req.GetScoreBase()
	t.GameCount = req.GetGameCount()
	t.PlayerCount = req.GetPlayerCount()
	t.GameConfig = req.GetGameConfig()
	return ack
}

func (t *Table) HandleAddPlayer(ctx context.Context, req *sproto.AddPlayerReq) *sproto.AddPlayerAck {
	ack := &sproto.AddPlayerAck{
		ErrorCode: int32(0), // 成功
	}
	if t.isOnTable(req.Playerid) {
		ack.ErrorCode = int32(1) // 玩家已在桌上
		return ack
	}

	player := playerManager.GetPlayer(req.Playerid)
	player.SetSeat(req.Seatnum)
	player.AddScore(req.Score)
	t.Players[req.Playerid] = player

	return ack
}

func (t *Table) HandleCancelTable(ctx context.Context, req *sproto.CancelTableReq) (ack *sproto.CancelTableAck) {
	ack = &sproto.CancelTableAck{
		ErrorCode: int32(0), // 成功
	}
	if t.Status == TableStatusPlaying {
		ack.ErrorCode = int32(1)
		return
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
	return ack
}

func (t *Table) SendToMatch() error {
	rsp := &cproto.CommonResponse{Err: cproto.ErrCode_OK}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := t.app.RPCTo(ctx, t.MatchServerId, "match.game.message", rsp, &sproto.Match2GameAck{}); err != nil {
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
	if m, err := t.app.SendPushToUsers("gamemsg", msg, players, "proxy"); err != nil {
		logrus.Errorf("send game message failed: %v", err)
	} else {
		logrus.Infof("broadcast game message to players: %v", m)
	}
}

func (t *Table) SendMsg(msg *cproto.GameAck, playerID string) error {
	if m, err := t.app.SendPushToUsers("gamemsg", msg, []string{playerID}, "proxy"); err != nil {
		logrus.Errorf("send game message to player %s failed: %v", playerID, err)
		return err
	} else {
		logrus.Infof("send game message to player %s: %v", playerID, m)
	}
	return nil
}
