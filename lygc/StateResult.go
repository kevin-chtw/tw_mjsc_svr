package lygc

import (
	"github.com/golang/protobuf/proto"
	"github.com/kevin-chtw/tw_game_svr/mahjong"
)

type StateResult struct {
	*State
	huSeats   []int
	paoSeat   mahjong.ISeatID
	paoCiSeat mahjong.ISeatID
	qiangGang bool
}

func NewStateResult(game mahjong.IGame) *StateResult {
	return &StateResult{
		State:     NewState(game),
		huSeats:   make([]int, 0),
		paoSeat:   mahjong.SeatNull,
		paoCiSeat: mahjong.SeatNull,
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

type StateResultLiuJu struct {
	*StateResult
}

func NewStateResultLiuJu(game mahjong.IGame) *StateResultLiuJu {
	return &StateResultLiuJu{
		StateResult: NewStateResult(game),
	}
}

func (s *StateResultLiuJu) OnEnter() {
	s.onPlayerLiuJu()
}

func (s *StateResultLiuJu) onPlayerLiuJu() {
	// 实现流局处理逻辑
}

type StateResultPaoHu struct {
	*StateResult
}

func NewStateResultPaoHu(huSeats []int, isGrabKon bool) *StateResultPaoHu {
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

type StateResultSelfHu struct {
	*StateResult
	dianCiSeat int
	huType     EHuType
}

func NewStateResultSelfHu(game mahjong.IGame, huType EHuType, seat int) *StateResultSelfHu {
	return &StateResultSelfHu{
		StateResult: NewStateResult(game),
		dianCiSeat:  seat,
		huType:      huType,
	}
}

func (s *StateResultSelfHu) OnEnter() {
	// 实现自胡状态进入逻辑
}
