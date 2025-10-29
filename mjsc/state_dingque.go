package mjsc

import (
	"errors"
	"time"

	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/kevin-chtw/tw_proto/game/pbsc"
	"google.golang.org/protobuf/proto"
)

type StateDingque struct {
	*State
}

func NewStateDingque(game mahjong.IGame, args ...any) mahjong.IState {
	s := &StateDingque{
		State: NewState(game),
	}
	return s
}

func (s *StateDingque) OnEnter() {
	s.game.sender.sendDingQueAck()
	s.AsyncMsgTimer(s.OnMsg, time.Second*time.Duration(8), s.OnTimeout)
}

func (s *StateDingque) OnMsg(seat int32, msg proto.Message) error {
	req, ok := msg.(*pbsc.SCDingQueReq)
	if !ok {
		return nil
	}
	if req == nil || !s.game.sender.IsRequestID(seat, req.Requestid) {
		return errors.New("invalid request")
	}

	if _, ok := s.game.play.queColors[seat]; ok {
		return errors.New("color already selected")
	}

	if req.Color < int32(mahjong.ColorCharacter) || req.Color > int32(mahjong.ColorDot) {
		return errors.New("invalid color")
	}

	s.game.play.queColors[seat] = mahjong.EColor(req.Color)
	s.game.sender.sendDingQueFinishAck(seat, req.Color)
	if len(s.game.play.queColors) >= int(s.game.GetPlayerCount()) {
		s.dingque()
	}
	return nil
}

func (s *StateDingque) OnTimeout() {
	if s.game.MatchType == "fdtable" {
		return
	}
	for i := int32(0); i < s.game.GetPlayerCount(); i++ {
		if _, ok := s.game.play.queColors[i]; !ok {
			s.game.play.queColors[i] = s.game.play.queRecommand(i)
		}
	}
	s.dingque()
}

func (s *StateDingque) dingque() {
	s.game.sender.sendDingQueResultAck(s.game.play.queColors)
	s.game.SetNextState(NewStateDiscard)
}
