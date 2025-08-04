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

// Message 处理匹配服务消息
func (m *MatchService) Message(ctx context.Context, req *sproto.Match2GameReq) {
	var err error
	if req.GetAddTableReq() != nil {
		err = m.handleAddTable(ctx, req)
	}
	if req.GetAddPlayerReq() != nil {
		err = m.handleAddPlayer(ctx, req)
	}

	if req.GetCancelTableReq() != nil {
		err = m.handleCancelTable(ctx, req)
	}

	if err != nil {
		logrus.Errorf("MatchService error handling message: %v, error: %v", req, err)
	}
}

// handleStartGame 处理开始游戏请求
func (m *MatchService) handleAddTable(ctx context.Context, req *sproto.Match2GameReq) error {
	table := game.GetTableManager().LoadOrStore(req.Matchid, req.Tableid)
	if table == nil {
		return fmt.Errorf("table not found for match %d, table %d", req.Matchid, req.Tableid)
	}
	if table.Status != game.TableStatusPreparing {
		return fmt.Errorf("table %d is not in preparing status", req.Tableid)
	}
	table.HandleStartGame(ctx, req.GetAddTableReq())
	return nil
}

func (m *MatchService) handleAddPlayer(ctx context.Context, req *sproto.Match2GameReq) error {
	table := game.GetTableManager().Get(req.Matchid, req.Tableid)
	if table == nil {
		return fmt.Errorf("table not found for match %d, table %d", req.Matchid, req.Tableid)
	}
	if table.Status != game.TableStatusPreparing {
		return fmt.Errorf("table %d is not in preparing status", req.Tableid)
	}

	table.HandleAddPlayer(ctx, req.GetAddPlayerReq())
	return nil
}

func (m *MatchService) handleCancelTable(ctx context.Context, req *sproto.Match2GameReq) error {
	table := game.GetTableManager().Get(req.Matchid, req.Tableid)
	if table == nil {
		return fmt.Errorf("table not found for match %d, table %d", req.Matchid, req.Tableid)
	}
	if table.Status != game.TableStatusPreparing {
		return fmt.Errorf("table %d is not in preparing status", req.Tableid)
	}

	table.HandleCancelTable(ctx, req.GetCancelTableReq())
	logrus.Infof("Game cancelled for match %d, table %d", req.Matchid, req.Tableid)
	return nil
}
