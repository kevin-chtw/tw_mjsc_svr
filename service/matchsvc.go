package service

import (
	"context"
	"fmt"

	"github.com/kevin-chtw/tw_game_svr/game"

	"github.com/kevin-chtw/tw_proto/sproto"
	"github.com/sirupsen/logrus"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/component"
)

// MatchService 独立的匹配服务
type MatchService struct {
	component.Base
	app pitaya.Pitaya
}

// NewMatchService 创建独立的匹配服务
func NewMatchService(app pitaya.Pitaya) *MatchService {
	return &MatchService{
		app: app,
	}
}

// Init 组件初始化
func (m *MatchService) Init() {
	logrus.Info("MatchService initialized")
}

// AfterInit 组件初始化后执行
func (m *MatchService) AfterInit() {
	logrus.Info("MatchService ready")
	// 初始化时可以加载持久化的match状态
}

// Message 处理匹配消息
func (m *MatchService) Message(ctx context.Context, req *sproto.Match2GameReq) {
	switch {
	case req.StartGameReq != nil:
		if err := m.handleStartGame(ctx, req.StartGameReq); err != nil {
			logrus.Errorf("Handle start game failed: %v", err)
		}
	case req.CancelMatchReq != nil:
		m.handleCancelMatch(ctx, req.CancelMatchReq)
	default:
		logrus.Warnf("Unknown match message type: %v", req)
	}
}

// handleStartGame 处理开始游戏请求
func (m *MatchService) handleStartGame(ctx context.Context, req *sproto.StartGameReq) error {
	table := game.GetTableManager().LoadOrStore(req.Matchid, req.Tableid)
	if table == nil {
		return fmt.Errorf("failed to create table for match %s and table %s", req.Matchid, req.Tableid)

	}

	table.HandleStartGame(ctx, req)
	return nil
}

// handleCancelMatch 处理取消匹配请求
func (m *MatchService) handleCancelMatch(ctx context.Context, req *sproto.CancelMatchReq) error {
	table := game.GetTableManager().Get(req.Matchid, req.Tableid)
	if table == nil {
		return fmt.Errorf("failed to create table for match %s and table %s", req.Matchid, req.Tableid)
	}

	table.HandleCancelGame(ctx, req)
	game.GetTableManager().Delete(req.Matchid, req.Tableid)
	return nil
}
