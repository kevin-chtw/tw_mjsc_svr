package sc

import (
	"github.com/kevin-chtw/tw_game_svr/mahjong"

	"github.com/golang/protobuf/proto"
)

type StateWait struct {
	*State
	operatesForSeats   [4]*mahjong.Operates // 每个座位可执行的操作
	reqOperateForSeats [4]int               // 每个座位已请求的操作
}

func NewStateWait(game mahjong.IGame) *StateWait {
	return &StateWait{
		State:              NewState(game),
		operatesForSeats:   [4]*mahjong.Operates{},
		reqOperateForSeats: [4]int{},
	}
}

func (s *StateWait) OnEnter() {
	// 实现等待状态进入逻辑
}

func (s *StateWait) OnReqRequest(seat int, req *proto.Message) {
	// 实现处理等待请求逻辑
}

func (s *StateWait) Timeout() {
	// 实现超时处理逻辑
}

func (s *StateWait) SetRequestAction(seat int, operate int) {
	if seat >= 0 && seat < len(s.reqOperateForSeats) {
		s.reqOperateForSeats[seat] = operate
	}
}

func (s *StateWait) TryHandleAction() {
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

func (s *StateWait) isValidOperate(seat int, operate int) bool {
	// 检查操作是否有效
	if seat < 0 || seat >= len(s.operatesForSeats) {
		return false
	}
	if s.operatesForSeats[seat] == nil {
		return false
	}
	return s.operatesForSeats[seat].HasOperate(int(operate))
}

func (s *StateWait) getMaxOperate(seat int) int {
	// 获取最高优先级的操作
	return mahjong.OperateNone
}

func (s *StateWait) getDefaultOperate(seat int) int {
	// 获取默认操作
	return mahjong.OperateNone
}
