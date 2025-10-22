package mjsc

import (
	"github.com/kevin-chtw/tw_common/game"
	"github.com/kevin-chtw/tw_common/mahjong"
	"github.com/kevin-chtw/tw_proto/game/pbsc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Sender struct {
	*mahjong.Sender
}

func NewSender(game *Game) *Sender {
	s := &Sender{}
	s.Sender = mahjong.NewSender(game.Game, game.play.Play, s)
	return s
}

func (m *Sender) PackMsg(msg proto.Message) (proto.Message, error) {
	data, err := anypb.New(msg)
	if err != nil {
		return nil, err
	}
	ack := &pbsc.SCAck{Ack: data}
	return ack, nil
}

func (s *Sender) sendSwapTilesAck() {
	ack := &pbsc.SCSwapTilesAck{
		Requestid: s.GetRequestID(game.SeatAll),
	}
	s.SendMsg(ack, game.SeatAll)
}

func (s *Sender) sendDingQueAck() {
	ack := &pbsc.SCDingQueAck{
		Requestid: s.GetRequestID(game.SeatAll),
	}
	s.SendMsg(ack, game.SeatAll)
}

func (s *Sender) sendSwapTilesResultAck(swapType int32, swaps []*pbsc.SCSwapTiles) {
	ack := &pbsc.SCCSwapTilesResultAck{
		SwapType:  swapType,
		SwapTiles: swaps,
	}
	s.SendMsg(ack, game.SeatAll)
}
