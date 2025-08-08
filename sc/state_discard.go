package sc

import (
	"time"

	"github.com/kevin-chtw/tw_game_svr/mahjong"
	"github.com/kevin-chtw/tw_proto/scproto"
	"google.golang.org/protobuf/proto"
)

type StateDiscard struct {
	*State
	operates *mahjong.Operates
}

func NewStateDiscard(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateDiscard{
		State: NewState(game),
	}
}

func (s *StateDiscard) OnEnter() {
	s.operates = s.GetPlay().FetchSelfOperates()
	s.GetMessager().sendRequestAck(s.GetPlay().GetCurSeat(), s.operates)
	discardTime := s.game.GetRule().GetValue(RuleDiscardTime)
	s.AsyncMsgTimer(s.OnMsg, time.Second*time.Duration(discardTime), s.OnTimeout)
}

func (s *StateDiscard) OnMsg(seat int32, msg proto.Message) {
	req := msg.(*scproto.SCReq)

	aniReq := req.GetScAnimationReq()
	if aniReq != nil && seat == aniReq.Seat && s.game.IsRequestID(seat, aniReq.Requestid) {
		s.game.Game.SetNextState(NewStateDiscard)
	}
}

func (s *StateDiscard) AutoOperate(isTimeout bool) {
	// 实现自动操作逻辑
}

func (s *StateDiscard) Discard(tile int) {
	// 实现弃牌逻辑
}

func (s *StateDiscard) Kon(tile int) {
	// 实现杠牌逻辑
}

func (s *StateDiscard) OnTimeout() {
	// 实现超时处理逻辑
}
