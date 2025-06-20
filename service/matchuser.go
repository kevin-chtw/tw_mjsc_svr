package service

import (
	"context"

	"github.com/kevin-chtw/tw_proto/sproto"
	"github.com/sirupsen/logrus"
)

type MatchUser struct {
	serverid string
}

func NewMatchUser(serverid string) *MatchUser {
	return &MatchUser{serverid: serverid}
}

func (mu *MatchUser) OnMatchMsg(ctx context.Context, req *sproto.Match2GameReq) {
	if req.StartGameReq != nil {
		// Handle the start game request
		if err := mu.handleStartGame(ctx, req.StartGameReq); err != nil {
			// Log the error or handle it accordingly
			return
		}
	}
}

func (mu *MatchUser) handleStartGame(ctx context.Context, req *sproto.StartGameReq) error {
	logrus.Infof("Handling start game request for server %s with match ID %s", mu.serverid, req.Matchid)
	return nil
}
