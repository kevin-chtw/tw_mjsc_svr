package main

import (
	"context"
	"strconv"
	"time"

	"github.com/kevin-chtw/tw_common/gamebase/game"
	"github.com/kevin-chtw/tw_common/utils"
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

	serverType := utils.MJSC + "_trainer"
	pitaya.SetLogger(utils.Logger(logrus.InfoLevel))

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
	time.Sleep(time.Second * 10)
	table := game.GetTableManager().LoadOrStore(1, 1)
	table.HandleAddTable(context.Background(), &sproto.AddTableReq{
		ScoreBase:   1,
		GameCount:   10000,
		PlayerCount: 4,
	})
	for i := range 4 {
		table.HandleAddPlayer(context.Background(), &sproto.AddPlayerReq{
			Playerid: strconv.Itoa(i + 1),
			Bot:      true,
			Seat:     int32(i),
		})
	}
}
