package mjsc

import (
	"errors"
	"time"

	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/kevin-chtw/tw_proto/game/pbmj"
	"google.golang.org/protobuf/proto"
)

type StateWait struct {
	*State
	operatesForSeats   []*mahjong.Operates // 每个座位可执行的操作
	reqOperateForSeats map[int32]int       // 每个座位已请求的操作
}

func NewStateWait(game mahjong.IGame, args ...any) mahjong.IState {
	s := &StateWait{
		State: NewState(game),
	}
	s.operatesForSeats = make([]*mahjong.Operates, s.game.GetPlayerCount())
	s.reqOperateForSeats = make(map[int32]int)
	return s
}

func (s *StateWait) OnEnter() {
	discardSeat := s.game.play.GetCurSeat()
	for i := int32(0); i < s.game.GetPlayerCount(); i++ {
		if i == discardSeat {
			continue
		}
		trusted := s.game.GetPlayer(i).IsTrusted()
		operates := s.game.play.FetchWaitOperates(i)
		s.operatesForSeats[i] = operates

		if operates.Value != mahjong.OperatePass && !trusted {
			s.game.sender.SendRequestAck(i, operates)
		} else {
			s.setReqOperate(i, s.getDefaultOperate(i))
		}
	}

	timeout := s.game.GetRule().GetValue(RuleWaitTime) + 1
	s.AsyncMsgTimer(s.OnMsg, time.Second*time.Duration(timeout), s.Timeout)
	s.tryHandleAction()
}

func (s *StateWait) OnMsg(seat int32, msg proto.Message) error {
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
	s.setReqOperate(seat, int(optReq.RequestType))
	return nil
}

func (s *StateWait) Timeout() {
	if s.game.MatchType == "fdtable" {
		return
	}
	for i := int32(0); i < s.game.GetPlayerCount(); i++ {
		if i == s.game.play.GetCurSeat() {
			continue
		}
		if _, ok := s.reqOperateForSeats[i]; !ok {
			s.setReqOperate(i, s.getDefaultOperate(i))
		}
	}
	s.tryHandleAction()
}

func (s *StateWait) setReqOperate(seat int32, operate int) {
	if s.game.IsValidSeat(seat) {
		s.reqOperateForSeats[seat] = operate
	}
}

func (s *StateWait) getReqOperate(seat int32) (int, bool) {
	operate, ok := s.reqOperateForSeats[seat]
	return operate, ok
}

func (s *StateWait) tryHandleAction() {
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
		return
	}

	maxOper := mahjong.OperatePass
	maxOperSeat := mahjong.SeatNull
	isMaxReq := true
	for i := int32(1); i < s.game.GetPlayerCount(); i++ {
		seat := mahjong.GetNextSeat(curSeat, i, s.game.GetPlayerCount())
		if operate, ok := s.getReqOperate(seat); ok {
			if operate > maxOper {
				maxOper = operate
				maxOperSeat = seat
				isMaxReq = true
			}
		} else if operate := s.getMaxOperate(seat); operate > maxOper {
			maxOper = operate
			maxOperSeat = seat
			isMaxReq = false
		}
	}
	if isMaxReq {
		s.excuteOperate(maxOperSeat, maxOper)
	}
}

func (s *StateWait) excuteOperate(seat int32, operate int) {
	if operate == mahjong.OperateKon {
		s.game.play.ZhiKon(seat)
		s.game.sender.SendKonAck(seat, s.game.play.GetCurTile(), mahjong.KonTypeZhi)
		s.toDrawState(seat)
		return
	}
	if operate == mahjong.OperatePon {
		s.game.play.Pon(seat)
		s.game.sender.SendPonAck(seat, s.game.play.GetCurTile())
		s.toDiscardState(seat)
		return
	}
	s.toDrawState(mahjong.SeatNull)
}

func (s *StateWait) excuteHu(huSeats []int32) {
	s.game.SetNextState(NewStatePaohu, huSeats)
}

func (s *StateWait) toDrawState(seat int32) {
	s.game.play.DoSwitchSeat(seat)
	s.game.SetNextState(NewStateDraw)
}

func (s *StateWait) toDiscardState(seat int32) {
	s.game.play.DoSwitchSeat(seat)
	s.game.SetNextState(NewStateDiscard)
}

func (s *StateWait) isValidOperate(seat int32, operate int) bool {
	// 检查操作是否有效
	if !s.game.IsValidSeat(seat) {
		return false
	}
	if s.operatesForSeats[seat] == nil {
		return false
	}
	return s.operatesForSeats[seat].HasOperate(int32(operate))
}

func (s *StateWait) getMaxOperate(seat int32) int {
	if ops := s.operatesForSeats[seat]; ops != nil {
		if ops.HasOperate(mahjong.OperateChowTing) {
			return mahjong.OperateChowTing
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

func (s *StateWait) getDefaultOperate(seat int32) int {
	ops := s.operatesForSeats[seat]
	if ops != nil && ops.HasOperate(mahjong.OperateHu) {
		return mahjong.OperateHu
	}
	return mahjong.OperatePass
}
