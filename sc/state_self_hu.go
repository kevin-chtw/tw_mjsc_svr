package sc

import (
	"time"

	"github.com/kevin-chtw/tw_game_svr/mahjong"
	"github.com/kevin-chtw/tw_proto/scproto"
	"google.golang.org/protobuf/proto"
)

type StateSelfHu struct {
	*StateResult
}

func NewStateSelfHu(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateSelfHu{
		StateResult: NewStateResult(game),
	}
}

func (s *StateSelfHu) OnEnter() {
	s.huSeats = append(s.huSeats, s.GetPlay().GetCurSeat())
	multiples := s.GetPlay().SelfHu()
	s.game.GetMessager().sendHuAck(s.huSeats, mahjong.SeatNull)
	s.game.GetScorelator().Calculate(multiples)

	s.game.GetMessager().sendAnimationAck()
	s.AsyncMsgTimer(s.OnMsg, time.Second*5, s.game.NotifyGameOver)
}

func (s *StateSelfHu) OnMsg(seat int32, msg proto.Message) {
	req := msg.(*scproto.SCReq)
	aniReq := req.GetScAnimationReq()
	if aniReq != nil && seat == aniReq.Seat && s.game.IsRequestID(seat, aniReq.Requestid) {
		s.game.NotifyGameOver()
	}
}
