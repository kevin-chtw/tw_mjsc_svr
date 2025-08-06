package service

import (
	"context"

	"github.com/kevin-chtw/tw_game_svr/game"
	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/sirupsen/logrus"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/component"
)

// PlayerService 独立的玩家服务
type PlayerService struct {
	component.Base
	app pitaya.Pitaya
}

// NewPlayerService 创建独立的玩家服务
func NewPlayerService(app pitaya.Pitaya) *PlayerService {
	return &PlayerService{
		app: app,
	}
}

func (p *PlayerService) Message(ctx context.Context, req *cproto.GameReq) {
	logrus.Infof("Received player message: %v", req)
	userID := p.app.GetSessionFromCtx(ctx).UID()
	if userID == "" {
		logrus.Error("Received player message with empty user ID")
		return
	}

	player := game.GetPlayerManager().GetPlayer(userID)
	player.HandleMessage(ctx, req)
}
