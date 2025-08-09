package sc

import (
	"time"

	"github.com/kevin-chtw/tw_game_svr/mahjong"

	"google.golang.org/protobuf/proto"
)

type StateWait struct {
	*State
	operatesForSeats   []*mahjong.Operates // 每个座位可执行的操作
	reqOperateForSeats []int               // 每个座位已请求的操作
}

func NewStateWait(game mahjong.IGame, args ...any) mahjong.IState {
	s := &StateWait{
		State: NewState(game),
	}
	s.operatesForSeats = make([]*mahjong.Operates, s.game.GetPlayerCount())
	s.reqOperateForSeats = make([]int, s.game.GetPlayerCount())
	return s
}

func (s *StateWait) OnEnter() {
	discardSeat := s.GetPlay().GetCurSeat()
	for i := int32(0); i < s.game.GetPlayerCount(); i++ {
		if i == discardSeat {
			continue
		}
		trusted := s.game.GetPlayer(i).IsTrusted()
		operates := s.GetPlay().FetchWaitOperates(i)
		s.operatesForSeats[i] = operates

		if operates.Value != 0 && !trusted {
			s.GetMessager().sendRequestAck(i, operates)
		} else {
			s.setRequestAction(i, s.getDefaultOperate(i))
		}
	}

	timeout := s.game.GetRule().GetValue(RuleWaitTime) + 1
	s.AsyncMsgTimer(s.OnPlayerMsg, time.Second*time.Duration(timeout), s.Timeout)
	s.tryHandleAction()
}

func (s *StateWait) OnPlayerMsg(seat int32, req proto.Message) {
	// 实现处理等待请求逻辑
}

func (s *StateWait) Timeout() {
	// 实现超时处理逻辑
	s.tryHandleAction()
}
func (s *StateWait) setRequestAction(seat int32, operate int) {
	if s.game.IsValidSeat(seat) {
		s.reqOperateForSeats[seat] = operate
	}
}

func (s *StateWait) tryHandleAction() {
	// 尝试处理所有玩家的操作
}

func (s *StateWait) ExcuteAction(seat int, operate int) {
	// 执行指定座位的操作
}

func (s *StateWait) ExcuteActionHu(huSeats []int) {
	// 执行胡牌操作
}

func (s *StateWait) ToDrawState(seat int) {
	// 转换到抽牌状态
}

func (s *StateWait) ToDiscardState(seat int) {
	// 转换到弃牌状态
}

func (s *StateWait) isValidOperate(seat int, operate int32) bool {
	// 检查操作是否有效
	if seat < 0 || seat >= len(s.operatesForSeats) {
		return false
	}
	if s.operatesForSeats[seat] == nil {
		return false
	}
	return s.operatesForSeats[seat].HasOperate(operate)
}

func (s *StateWait) getMaxOperate(seat int32) int {
	// 获取最高优先级的操作
	return mahjong.OperateNone
}

func (s *StateWait) getDefaultOperate(seat int32) int {
	// 获取默认操作
	return mahjong.OperateNone
}
