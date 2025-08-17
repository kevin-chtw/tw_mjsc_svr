package main

import (
	"strings"

	"github.com/kevin-chtw/tw_common/utils"
	"github.com/kevin-chtw/tw_game_svr/game"
	"github.com/kevin-chtw/tw_game_svr/sc"
	"github.com/kevin-chtw/tw_game_svr/service"
	"github.com/sirupsen/logrus"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/component"
	"github.com/topfreegames/pitaya/v3/pkg/config"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
)

var app pitaya.Pitaya

func main() {
	serverType := "game"
	pitaya.SetLogger(utils.Logger(logrus.InfoLevel))

	config := config.NewDefaultPitayaConfig()
	builder := pitaya.NewDefaultBuilder(false, serverType, pitaya.Cluster, map[string]string{}, *config)
	app = builder.Build()

	defer app.Shutdown()

	logger.Log.Infof("Pitaya server of type %s started", serverType)
	game.Register(utils.GameID_MahjongSC, sc.NewGame)
	game.InitGame(app)
	initServices()
	app.Start()
}

func initServices() {
	matchsvc := service.NewMatch(app)
	app.Register(matchsvc, component.WithName("match"), component.WithNameFunc(strings.ToLower))
	app.RegisterRemote(matchsvc, component.WithName("match"), component.WithNameFunc(strings.ToLower))

	playersvc := service.NewPlayer(app)
	app.Register(playersvc, component.WithName("player"), component.WithNameFunc(strings.ToLower))
	app.RegisterRemote(playersvc, component.WithName("player"), component.WithNameFunc(strings.ToLower))
}
