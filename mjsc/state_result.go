package mjsc

import (
	"github.com/kevin-chtw/tw_common/mahjong"
	"github.com/kevin-chtw/tw_proto/scproto"
	"google.golang.org/protobuf/proto"
)

type StateResult struct {
	*State
	huSeats []int32
}

func NewStateResult(game mahjong.IGame) *StateResult {
	return &StateResult{
		State:   NewState(game),
		huSeats: make([]int32, 0),
	}
}

func (s *StateResult) onMsg(seat int32, msg proto.Message) error {
	req := msg.(*scproto.SCReq)
	aniReq := req.GetScAnimationReq()
	if aniReq != nil && seat == aniReq.Seat && s.game.IsRequestID(seat, aniReq.Requestid) {
		s.game.OnGameOver()
	}
	return nil
}
