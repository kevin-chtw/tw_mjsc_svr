package main

import (
	"strings"

	"github.com/kevin-chtw/tw_common/gamebase/game"
	"github.com/kevin-chtw/tw_common/gamebase/service"
	"github.com/kevin-chtw/tw_common/utils"
	"github.com/kevin-chtw/tw_mjsc_svr/ai"
	"github.com/kevin-chtw/tw_mjsc_svr/bot"
	"github.com/kevin-chtw/tw_mjsc_svr/mjsc"
	"github.com/sirupsen/logrus"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/component"
	"github.com/topfreegames/pitaya/v3/pkg/config"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"github.com/topfreegames/pitaya/v3/pkg/serialize"
)

var app pitaya.Pitaya

func main() {
	// 关闭训练模式（仅推理）
	ai.SetTrainingMode(false)

	serverType := utils.MJSC
	pitaya.SetLogger(utils.Logger(logrus.InfoLevel))

	config := config.NewDefaultPitayaConfig()
	config.SerializerType = uint16(serialize.PROTOBUF)
	config.Handler.Messages.Compression = false
	builder := pitaya.NewDefaultBuilder(false, serverType, pitaya.Cluster, map[string]string{}, *config)
	app = builder.Build()

	defer app.Shutdown()

	logger.Log.Infof("Pitaya server of type %s started", serverType)
	game.Init(app, mjsc.NewGame, bot.NewPlayer)
	initServices()
	app.Start()
}

func initServices() {
	remote := service.NewRemote(app)
	app.RegisterRemote(remote, component.WithName("remote"), component.WithNameFunc(strings.ToLower))

	playersvc := service.NewPlayer(app)
	app.Register(playersvc, component.WithName("player"), component.WithNameFunc(strings.ToLower))
}
