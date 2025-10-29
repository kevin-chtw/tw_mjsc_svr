package mjsc

import (
	"errors"
	"time"

	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/kevin-chtw/tw_proto/game/pbmj"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"google.golang.org/protobuf/proto"
)

type StateDiscard struct {
	*State
	operates *mahjong.Operates
	handlers map[int32]func(tile mahjong.Tile)
}

func NewStateDiscard(game mahjong.IGame, args ...any) mahjong.IState {
	s := &StateDiscard{
		State:    NewState(game),
		handlers: make(map[int32]func(tile mahjong.Tile)),
	}
	s.handlers[mahjong.OperateDiscard] = s.discard
	s.handlers[mahjong.OperateKon] = s.kon
	s.handlers[mahjong.OperateHu] = s.hu
	return s
}

func (s *StateDiscard) OnEnter() {
	s.operates = s.game.play.FetchSelfOperates()
	s.game.sender.SendRequestAck(s.game.play.GetCurSeat(), s.operates)
	discardTime := s.game.GetRule().GetValue(RuleDiscardTime)
	logger.Log.Infof("discard time: %d", discardTime)
	s.AsyncMsgTimer(s.OnMsg, time.Second*time.Duration(discardTime), s.OnTimeout)
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

func (s *StateDiscard) discard(tile mahjong.Tile) {
	if s.game.play.discard(tile) {
		s.game.sender.SendDiscardAck()
		s.game.SetNextState(NewStateWait)
	}
}

func (s *StateDiscard) kon(tile mahjong.Tile) {
	if s.game.play.TryKon(tile, mahjong.KonTypeBu) {
		s.game.sender.SendKonAck(s.game.play.GetCurSeat(), tile, mahjong.KonTypeBu)
		s.game.SetNextState(NewStateAfterBukon)
	} else if s.game.play.TryKon(tile, mahjong.KonTypeAn) {
		s.game.sender.SendKonAck(s.game.play.GetCurSeat(), tile, mahjong.KonTypeAn)
		scores := s.game.scorelator.Check(s.game.play.GetCurSeat(), mahjong.SeatNull, -2, -2)
		s.game.sender.SendScoreChangeAck(mahjong.ScoreReasonAnKon, scores, s.game.play.GetCurTile(), mahjong.SeatNull, nil)
		s.game.SetNextState(NewStateDraw)
	}
}

func (s *StateDiscard) hu(tile mahjong.Tile) {
	s.game.SetNextState(NewStateZimo)
}

func (s *StateDiscard) OnTimeout() {
	if s.game.MatchType == "fdtable" {
		return
	}
	logger.Log.Warnf("discard timeout")
	s.discard(mahjong.TileNull)
}
