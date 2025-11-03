package mjsc

import (
	"errors"
	"time"

	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/kevin-chtw/tw_proto/game/pbmj"
	"google.golang.org/protobuf/proto"
)

type StateAfterBukon struct {
	*State
	operatesForSeats   []*mahjong.Operates // 每个座位可执行的操作
	reqOperateForSeats map[int32]int       // 每个座位已请求的操作
}

func NewStateAfterBukon(game mahjong.IGame, args ...any) mahjong.IState {
	s := &StateAfterBukon{
		State: NewState(game),
	}
	s.operatesForSeats = make([]*mahjong.Operates, s.game.GetPlayerCount())
	s.reqOperateForSeats = make(map[int32]int)
	return s
}

func (s *StateAfterBukon) OnEnter() {
	konSeat := s.game.play.GetCurSeat()
	for i := int32(0); i < s.game.GetPlayerCount(); i++ {
		if i == konSeat || s.game.GetPlayer(i).IsOut() {
			continue
		}

		operates := s.game.play.FetchAfterBuKonOperates(i, newCheckerPao(s.game.play))
		s.operatesForSeats[i] = operates

		if operates.Value != mahjong.OperatePass && !s.game.GetPlayer(i).IsTrusted() {
			s.game.sender.SendRequestAck(i, operates)
		} else {
			s.reqOperateForSeats[i] = mahjong.OperatePass
		}
	}
	timeout := s.game.GetRule().GetValue(RuleWaitTime) + 1
	s.AsyncMsgTimer(s.OnMsg, time.Second*time.Duration(timeout), s.OnTimeout)
	s.tryHandleAction()
}

func (s *StateAfterBukon) OnMsg(seat int32, msg proto.Message) error {
	optReq, ok := msg.(*pbmj.MJRequestReq)
	if !ok {
		return nil
	}
	if optReq == nil || optReq.Seat != seat || !s.game.sender.IsRequestID(seat, optReq.Requestid) {
		return errors.New("invalid request")
	}

	if !s.isValidOperate(seat, int(optReq.RequestType)) {
		return errors.New("invalid operate")
	}
	s.reqOperateForSeats[seat] = int(optReq.RequestType)
	s.tryHandleAction()
	return nil
}

func (s *StateAfterBukon) isValidOperate(seat int32, operate int) bool {
	// 检查操作是否有效
	if !s.game.IsValidSeat(seat) {
		return false
	}
	if s.operatesForSeats[seat] == nil {
		return false
	}
	return s.operatesForSeats[seat].HasOperate(int32(operate))
}

func (s *StateAfterBukon) tryHandleAction() {
	curSeat := s.game.play.GetCurSeat()
	huSeats := make([]int32, 0)
	for i := int32(1); i < s.game.GetPlayerCount(); i++ {
		seat := mahjong.GetNextSeat(curSeat, i, s.game.GetPlayerCount())
		if operate, ok := s.getReqOperate(seat); ok {
			if operate == mahjong.OperateHu {
				huSeats = append(huSeats, seat)
			}
		} else if s.getMaxOperate(seat) == mahjong.OperateHu {
			return
		}
	}

	if len(huSeats) > 0 {
		s.excuteHu(huSeats)
	} else {
		scores := s.game.scorelator.CalcKon(mahjong.ScoreReasonBuKon, s.game.play.GetCurSeat(), mahjong.SeatNull, 1, 1)
		s.game.sender.SendScoreChangeAck(mahjong.ScoreReasonBuKon, scores, s.game.play.GetCurTile(), mahjong.SeatNull, nil)
		s.game.SetNextState(NewStateDraw)
	}
}

func (s *StateAfterBukon) excuteHu(huSeats []int32) {
	multiples := s.game.play.PaoHu(huSeats)
	scores := s.game.scorelator.CalcMulti(mahjong.SeatNull, mahjong.ScoreReasonHu, multiples)
	s.game.sender.SendHuAck(huSeats, s.game.play.GetCurSeat())
	s.game.sender.SendScoreChangeAck(mahjong.ScoreReasonHu, scores, s.game.play.GetCurTile(), s.game.play.GetCurSeat(), huSeats)
	for _, seat := range huSeats {
		s.game.GetPlayer(seat).SetOut()
	}
	nextSeat := mahjong.GetNextSeat(huSeats[len(huSeats)-1], 1, s.game.GetPlayerCount())
	s.game.play.DoSwitchSeat(nextSeat)
	s.game.SetNextState(NewStateDraw)
}

func (s *StateAfterBukon) getReqOperate(seat int32) (int, bool) {
	operate, ok := s.reqOperateForSeats[seat]
	return operate, ok
}

func (s *StateAfterBukon) getMaxOperate(seat int32) int {
	if ops := s.operatesForSeats[seat]; ops != nil {
		if ops.HasOperate(mahjong.OperateHu) {
			return mahjong.OperateHu
		}
		if ops.HasOperate(mahjong.OperateKon) {
			return mahjong.OperateKon
		}
		if ops.HasOperate(mahjong.OperatePon) {
			return mahjong.OperatePon
		}
	}
	return mahjong.OperatePass
}

func (s *StateAfterBukon) OnTimeout() {
	if s.game.MatchType == "fdtable" {
		return
	}
	for i := int32(0); i < s.game.GetPlayerCount(); i++ {
		if i == s.game.play.GetCurSeat() {
			continue
		}
		if _, ok := s.reqOperateForSeats[i]; !ok {
			s.reqOperateForSeats[i] = mahjong.OperatePass
		}
	}
	s.tryHandleAction()
}
