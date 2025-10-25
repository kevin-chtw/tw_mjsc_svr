package mjsc

import (
	"github.com/kevin-chtw/tw_common/gamebase/game"
	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
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
	ack := &pbsc.SCSwapTilesResultAck{
		SwapType:  swapType,
		SwapTiles: swaps,
	}
	s.SendMsg(ack, game.SeatAll)
}

func (s *Sender) sendDingQueResultAck(queColors map[int32]mahjong.EColor) {
	ack := &pbsc.SCDingQueResultAck{
		Colors: make([]int32, 0),
	}
	for i := 0; i < len(queColors); i++ {
		ack.Colors = append(ack.Colors, int32(queColors[int32(i)]))
	}
	s.SendMsg(ack, game.SeatAll)
}
