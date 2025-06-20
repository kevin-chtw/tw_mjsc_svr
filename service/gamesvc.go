package service

import (
	"context"
	"sync"

	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/kevin-chtw/tw_proto/sproto"
	"github.com/sirupsen/logrus"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/component"
)

type GameSvc struct {
	component.Base
	app        pitaya.Pitaya
	matchusers sync.Map
}

func NewGameSvc(app pitaya.Pitaya) *GameSvc {
	return &GameSvc{app: app, matchusers: sync.Map{}}
}

func (s *GameSvc) MatchMsg(ctx context.Context, req *sproto.Match2GameReq) (*cproto.CommonResponse, error) {
	logrus.Debugf("MatchMsg: %v", req)
	matchuser := NewMatchUser(req.Serverid)
	if loaded, ok := s.matchusers.LoadOrStore(req.Serverid, matchuser); ok {
		matchuser = loaded.(*MatchUser)
	}

	matchuser.OnMatchMsg(ctx, req)
	return &cproto.CommonResponse{Err: cproto.ErrCode_OK}, nil
}
