package main

import (
	"context"
	"strconv"
	"time"

	"github.com/kevin-chtw/tw_common/gamebase/game"
	"github.com/kevin-chtw/tw_common/utils"
	"github.com/kevin-chtw/tw_mjsc_svr/ai"
	"github.com/kevin-chtw/tw_mjsc_svr/bot"
	"github.com/kevin-chtw/tw_mjsc_svr/mjsc"
	"github.com/kevin-chtw/tw_proto/sproto"
	"github.com/sirupsen/logrus"
	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	"github.com/topfreegames/pitaya/v3/pkg/config"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"github.com/topfreegames/pitaya/v3/pkg/serialize"
)

var app pitaya.Pitaya

func main() {
	// 开启训练模式
	ai.SetTrainingMode(true)

	// 初始化 Python AI 服务客户端
	if err := ai.InitHTTPAIClient("localhost:50051"); err != nil {
		logger.Log.Fatalf("Failed to init AI client: %v", err)
	}
	defer ai.GetHTTPAIClient().Close()

	serverType := utils.MJSC + "_trainer"
	pitaya.SetLogger(utils.Logger(logrus.WarnLevel))

	config := config.NewDefaultPitayaConfig()
	config.SerializerType = uint16(serialize.PROTOBUF)
	config.Handler.Messages.Compression = false
	builder := pitaya.NewDefaultBuilder(false, serverType, pitaya.Cluster, map[string]string{}, *config)
	app = builder.Build()

	defer app.Shutdown()

	logger.Log.Infof("Pitaya server of type %s started", serverType)
	game.Init(app, mjsc.NewGame, bot.NewPlayer)
	go train()
	app.Start()
}

func train() {
	time.Sleep(time.Second)
	table := game.GetTableManager().LoadOrStore(1, 1)
	table.HandleAddTable(context.Background(), &sproto.AddTableReq{
		ScoreBase:   1,
		GameCount:   20000,
		PlayerCount: 4,
		MatchType:   "trainer",
	})
	for i := range 4 {
		table.HandleAddPlayer(context.Background(), &sproto.AddPlayerReq{
			Playerid: strconv.Itoa(i + 1),
			Bot:      true,
			Seat:     int32(i),
		})
	}
}
