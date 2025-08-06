package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/kevin-chtw/tw_game_svr/game"

	"github.com/kevin-chtw/tw_proto/sproto"
	"github.com/sirupsen/logrus"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/component"
)

// MatchService 独立的匹配服务
type MatchService struct {
	component.Base
	app      pitaya.Pitaya
	handlers map[reflect.Type]func(context.Context, *sproto.Match2GameReq) (*sproto.Match2GameAck, error)
}

// NewMatchService 创建独立的匹配服务
func NewMatchService(app pitaya.Pitaya) *MatchService {
	return &MatchService{
		app:      app,
		handlers: make(map[reflect.Type]func(context.Context, *sproto.Match2GameReq) (*sproto.Match2GameAck, error)),
	}
}

// Init 组件初始化
func (m *MatchService) Init() {
	logrus.Info("MatchService initialized")
	m.handlers[reflect.TypeOf(&sproto.Match2GameReq_AddTableReq{})] = m.handleAddTable
	m.handlers[reflect.TypeOf(&sproto.Match2GameReq_AddPlayerReq{})] = m.handleAddPlayer
	m.handlers[reflect.TypeOf(&sproto.Match2GameReq_CancelTableReq{})] = m.handleCancelTable
	logrus.Info("MatchService handlers initialized")
}

// AfterInit 组件初始化后执行
func (m *MatchService) AfterInit() {
	logrus.Info("MatchService ready")
	// 初始化时可以加载持久化的match状态
}

// Message 处理匹配服务消息
func (m *MatchService) Message(ctx context.Context, req *sproto.Match2GameReq) (*sproto.Match2GameAck, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}

	if req.Req == nil {
		return nil, errors.New("empty oneof")
	}

	// 查找对应的handler
	fn, ok := m.handlers[reflect.TypeOf(req.Req)]
	if !ok {
		return nil, fmt.Errorf("no handler for message type: %T", req.Req)
	}

	return fn(ctx, req)
}

func (m *MatchService) NewMatch2GameAck(req *sproto.Match2GameReq) *sproto.Match2GameAck {
	return &sproto.Match2GameAck{
		Gameid:  0,
		Matchid: req.GetMatchid(),
		Tableid: req.GetTableid(),
	}
}

// handleStartGame 处理开始游戏请求
func (m *MatchService) handleAddTable(ctx context.Context, req *sproto.Match2GameReq) (*sproto.Match2GameAck, error) {
	table := game.GetTableManager().LoadOrStore(req.Matchid, req.Tableid)
	if table == nil {
		return nil, fmt.Errorf("table not found for match %d, table %d", req.Matchid, req.Tableid)
	}
	rsp := table.HandleAddTable(ctx, req.GetAddTableReq())
	ack := m.NewMatch2GameAck(req)
	ack.Ack = &sproto.Match2GameAck_AddTableAck{
		AddTableAck: rsp,
	}
	return ack, nil
}

func (m *MatchService) handleAddPlayer(ctx context.Context, req *sproto.Match2GameReq) (*sproto.Match2GameAck, error) {
	table := game.GetTableManager().Get(req.Matchid, req.Tableid)
	if table == nil {
		return nil, fmt.Errorf("table not found for match %d, table %d", req.Matchid, req.Tableid)
	}

	rsp := table.HandleAddPlayer(ctx, req.GetAddPlayerReq())
	ack := m.NewMatch2GameAck(req)
	ack.Ack = &sproto.Match2GameAck_AddPlayerAck{
		AddPlayerAck: rsp,
	}
	return ack, nil
}

func (m *MatchService) handleCancelTable(ctx context.Context, req *sproto.Match2GameReq) (*sproto.Match2GameAck, error) {
	table := game.GetTableManager().Get(req.Matchid, req.Tableid)
	if table == nil {
		return nil, fmt.Errorf("table not found for match %d, table %d", req.Matchid, req.Tableid)
	}

	rsp := table.HandleCancelTable(ctx, req.GetCancelTableReq())
	ack := m.NewMatch2GameAck(req)
	ack.Ack = &sproto.Match2GameAck_CancelTableAck{
		CancelTableAck: rsp,
	}
	return ack, nil
}
