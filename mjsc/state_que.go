package mjsc

import (
	"errors"

	"github.com/kevin-chtw/tw_common/mahjong"
	"github.com/kevin-chtw/tw_proto/game/pbmj"
	"google.golang.org/protobuf/proto"
)

type StateQue struct {
	*State
}

func NewStateQue(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateQue{
		State: NewState(game),
	}
}

func (s *StateQue) OnEnter() {
	s.game.sender.sendDingQueAck()

	s.WaitAni(func() { s.game.SetNextState(NewStateDiscard) })
}

func (s *StateDiscard) OnMsg(seat int32, msg proto.Message) error {
	if seat != s.game.play.GetCurSeat() {
		return errors.New("invalid seat")
	}

	optReq, ok := msg.(*pbmj.MJRequestReq)
	if !ok {
		return nil
	}
	if optReq == nil || optReq.Seat != seat || !s.game.sender.IsRequestID(seat, optReq.Requestid) {
		return errors.New("invalid request")
	}

	if !s.operates.HasOperate(optReq.RequestType) {
		return errors.New("invalid operate")
	}
	if handler, exists := s.handlers[optReq.RequestType]; exists {
		handler(mahjong.Tile(optReq.Tile))
	}
	return nil
}
