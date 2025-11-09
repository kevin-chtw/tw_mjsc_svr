package bot

import (
	"github.com/kevin-chtw/tw_common/gamebase/game"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"google.golang.org/protobuf/proto"
)

type Player struct {
	*game.BotPlayer
}

func NewPlayer(uid string) *game.BotPlayer {
	p := &Player{
		BotPlayer: game.NewBotPlayer(uid),
	}
	p.Bot = p
	return p.BotPlayer
}

func (p *Player) OnBotMsg(msg proto.Message) error {
	logger.Log.Info("bot player %s received msg: %v", p.Uid, msg)
	return nil
}
