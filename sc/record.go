package sc

import (
	"github.com/golang/protobuf/proto"
	"github.com/kevin-chtw/tw_game_svr/mahjong"
)

type Record struct {
	*mahjong.Record
}

func NewRecord(game *Game) *Record {
	return &Record{
		Record: mahjong.NewRecord(game.Game),
	}
}

func (r *Record) RecordAction(ack proto.Message, seat int) {
}
