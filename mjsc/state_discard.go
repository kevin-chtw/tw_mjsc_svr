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
	if s.game.GetPlayer(s.game.play.GetCurSeat()).IsTrusted() {
		s.discard(mahjong.TileNull)
		return
	}
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
		konType := mahjong.KonTypeBu
		if tile != s.game.play.GetCurTile() {
			konType = mahjong.KonTypeBa
		}
		s.game.sender.SendKonAck(s.game.play.GetCurSeat(), tile, konType)
		s.game.SetNextState(NewStateAfterBukon, konType)
	} else if s.game.play.TryKon(tile, mahjong.KonTypeAn) {
		s.game.sender.SendKonAck(s.game.play.GetCurSeat(), tile, mahjong.KonTypeAn)
		scores := s.game.scorelator.CalcKon(mahjong.ScoreReasonAnKon, s.game.play.GetCurSeat(), mahjong.SeatNull, 2, 2)
		s.game.sender.SendScoreChangeAck(mahjong.ScoreReasonAnKon, scores, s.game.play.GetCurTile(), mahjong.SeatNull, nil)
		s.game.SetNextState(NewStateDraw)
	}
}

func (s *StateDiscard) hu(tile mahjong.Tile) {
	huSeats := make([]int32, 0)
	huSeats = append(huSeats, s.game.play.GetCurSeat())
	var multiples []int64
	paoSeat := s.game.play.IsAfterZhiKon()
	if s.game.GetRule().GetValue(RuleDianKHSDP) != 0 && paoSeat != mahjong.SeatNull {
		multiples = s.game.play.DianKonHua(paoSeat)
	} else {
		multiples = s.game.play.Zimo()
	}
	s.game.sender.SendHuAck(huSeats, paoSeat)
	scores := s.game.scorelator.CalcMulti(s.game.play.GetCurSeat(), mahjong.ScoreReasonHu, multiples)
	s.game.sender.SendScoreChangeAck(mahjong.ScoreReasonHu, scores, s.game.play.GetCurTile(), paoSeat, huSeats)
	s.game.GetPlayer(s.game.play.GetCurSeat()).SetOut()
	s.game.play.DoSwitchSeat(mahjong.SeatNull)
	s.game.SetNextState(NewStateDraw)
}

func (s *StateDiscard) OnTimeout() {
	if s.game.MatchType == "fdtable" {
		return
	}
	logger.Log.Warnf("discard timeout")
	s.discard(mahjong.TileNull)
	s.game.sender.SendTrustAck(s.game.play.GetCurSeat(), true)
}
