package sc

import (
	"time"

	"github.com/kevin-chtw/tw_common/mahjong"
	"github.com/kevin-chtw/tw_proto/scproto"
	"google.golang.org/protobuf/proto"
)

type StateDeal struct {
	*State
}

var Deal = StateDeal{}

func NewStateDeal(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateDeal{
		State: NewState(game),
	}
}

func (s *StateDeal) OnEnter() {
	s.GetPlay().Deal()

	s.GetMessager().sendOpenDoorAck()
	s.GetMessager().sendAnimationAck()
	s.AsyncMsgTimer(s.OnMsg, time.Second*5, func() { s.game.SetNextState(NewStateDiscard) })
}

func (s *StateDeal) OnMsg(seat int32, msg proto.Message) {
	req := msg.(*scproto.SCReq)

	aniReq := req.GetScAnimationReq()
	if aniReq != nil && seat == aniReq.Seat && s.game.IsRequestID(seat, aniReq.Requestid) {
		s.game.SetNextState(NewStateDiscard)
	}
}
