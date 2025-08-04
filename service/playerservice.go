package service

import (
	"context"
	"errors"

	"github.com/kevin-chtw/tw_game_svr/game"
	"github.com/kevin-chtw/tw_proto/cproto"
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

func (p *PlayerService) Message(ctx context.Context, req *cproto.GameReq) (*cproto.CommonResponse, error) {
	userID := p.app.GetSessionFromCtx(ctx).UID()
	if userID == "" {
		return nil, errors.New("user ID is empty")
	}

	player := game.GetPlayerManager().GetPlayer(userID)
	player.HandleMessage(ctx, req)
	return &cproto.CommonResponse{
		Err: 0,
	}, nil
}
