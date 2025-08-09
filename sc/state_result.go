package sc

import (
	"github.com/kevin-chtw/tw_game_svr/mahjong"
	"google.golang.org/protobuf/proto"
)

type StateResult struct {
	*State
	huSeats   []int32
	paoSeat   int32
	qiangGang bool
}

func NewStateResult(game mahjong.IGame) *StateResult {
	return &StateResult{
		State:   NewState(game),
		huSeats: make([]int32, 0),
		paoSeat: mahjong.SeatNull,
	}
}

func (s *StateResult) OnReqRequest(seat int, req *proto.Message) {
	// 实现处理结果请求逻辑
}

func (s *StateResult) OnGameEnd(isLiuju bool) {
	// 实现游戏结束处理逻辑
}

func (s *StateResult) WaitClientAnimateDone() {
	// 等待客户端动画完成
	s.PrepareNextState()
}

func (s *StateResult) PrepareNextState() {
	// 准备进入下一个状态
}

func (s *StateResult) onPlayerHu(multiples []int64, paoSeat int, huType EHuType) {
	// 实现玩家胡牌处理逻辑
}

type StateResultPaoHu struct {
	*StateResult
}

func NewStateResultPaoHu(huSeats []int32, isGrabKon bool) *StateResultPaoHu {
	return &StateResultPaoHu{
		StateResult: &StateResult{
			huSeats:   huSeats,
			qiangGang: isGrabKon,
		},
	}
}

func (s *StateResultPaoHu) OnEnter() {
	// 实现跑胡状态进入逻辑
}
