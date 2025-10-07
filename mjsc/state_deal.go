package mjsc

import (
	"time"

	"github.com/kevin-chtw/tw_common/mahjong"
	"github.com/kevin-chtw/tw_proto/game/pbmj"
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
	s.game.play.Deal()

	s.game.sender.SendOpenDoorAck()
	s.game.sender.SendAnimationAck()
	s.AsyncMsgTimer(s.OnMsg, time.Second*5, func() { s.game.SetNextState(NewStateDiscard) })
}

func (s *StateDeal) OnMsg(seat int32, msg proto.Message) error {
	aniReq, ok := msg.(*pbmj.MJAnimationReq)
	if !ok {
		return nil
	}
	if aniReq != nil && seat == aniReq.Seat && s.game.IsRequestID(seat, aniReq.Requestid) {
		s.game.SetNextState(NewStateDiscard)
	}
	return nil
}
