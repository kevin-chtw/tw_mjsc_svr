package mjsc

import (
	"errors"
	"time"

	"github.com/kevin-chtw/tw_common/mahjong"
	"github.com/kevin-chtw/tw_proto/scproto"
	"google.golang.org/protobuf/proto"
)

type StateDiscard struct {
	*State
	operates *mahjong.Operates
	handlers map[int32]func(tile int32)
}

func NewStateDiscard(game mahjong.IGame, args ...any) mahjong.IState {
	s := &StateDiscard{
		State:    NewState(game),
		handlers: make(map[int32]func(tile int32)),
	}
	s.handlers[mahjong.OperateDiscard] = s.discard
	s.handlers[mahjong.OperateKon] = s.kon
	s.handlers[mahjong.OperateHu] = s.hu
	return s
}

func (s *StateDiscard) OnEnter() {
	s.operates = s.GetPlay().FetchSelfOperates()
	s.GetMessager().sendRequestAck(s.GetPlay().GetCurSeat(), s.operates)
	discardTime := s.game.GetRule().GetValue(RuleDiscardTime)
	s.AsyncMsgTimer(s.OnMsg, time.Second*time.Duration(discardTime), s.OnTimeout)
}

func (s *StateDiscard) OnMsg(seat int32, msg proto.Message) error {
	if seat != s.GetPlay().GetCurSeat() {
		return errors.New("invalid seat")
	}

	req := msg.(*scproto.SCReq)
	optReq := req.GetScRequestReq()
	if optReq == nil || optReq.Seat != seat || !s.game.IsRequestID(seat, optReq.Requestid) {
		return errors.New("invalid request")
	}

	if !s.operates.HasOperate(optReq.RequestType) {
		return errors.New("invalid operate")
	}
	if handler, exists := s.handlers[optReq.RequestType]; exists {
		handler(optReq.Tile)
	}
	return nil
}

func (s *StateDiscard) discard(tile int32) {
	s.GetPlay().Discard(tile)
	s.GetMessager().sendDiscardAck()
	s.game.SetNextState(NewStateWait)
}

func (s *StateDiscard) kon(tile int32) {
	if s.GetPlay().TryKon(tile, mahjong.KonTypeBu) {
		s.GetMessager().sendKonAck(s.GetPlay().GetCurSeat(), tile, mahjong.KonTypeBu)
		s.game.SetNextState(NewStateWait)
	} else if s.GetPlay().TryKon(tile, mahjong.KonTypeAn) {
		s.GetMessager().sendKonAck(s.GetPlay().GetCurSeat(), tile, mahjong.KonTypeAn)
		s.game.SetNextState(NewStateDraw)
	}
}

func (s *StateDiscard) hu(tile int32) {
	s.game.SetNextState(NewStateZimo)
}

func (s *StateDiscard) OnTimeout() {
	// 实现超时处理逻辑
}
