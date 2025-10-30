package mjsc

import (
	"errors"
	"math/rand"
	"time"

	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/kevin-chtw/tw_proto/game/pbsc"
	"google.golang.org/protobuf/proto"
)

type StateSwapTiles struct {
	*State
	swapTiles []*pbsc.SCSwapTiles
}

func NewStateSwapTiles(game mahjong.IGame, args ...any) mahjong.IState {
	s := &StateSwapTiles{
		State: NewState(game),
	}
	s.swapTiles = make([]*pbsc.SCSwapTiles, s.game.GetPlayerCount())
	return s
}

func (s *StateSwapTiles) OnEnter() {
	s.game.sender.sendSwapTilesAck()
	s.AsyncMsgTimer(s.OnMsg, time.Second*time.Duration(8), s.OnTimeout)
}

func (s *StateSwapTiles) OnMsg(seat int32, msg proto.Message) error {
	optReq, ok := msg.(*pbsc.SCSwapTilesReq)
	if !ok {
		return nil
	}
	if optReq == nil || !s.game.sender.IsRequestID(seat, optReq.Requestid) {
		return errors.New("invalid request")
	}

	if s.swapTiles[seat] != nil {
		return errors.New("already swapped")
	}

	if !s.game.play.GetPlayData(seat).CanExchangeOut(mahjong.Int32Tile(optReq.Tiles)) {
		return errors.New("invalid tiles")
	}
	s.swapTiles[seat] = &pbsc.SCSwapTiles{
		From:  seat,
		Tiles: optReq.Tiles,
	}

	// 检查是否所有玩家都已换牌
	s.game.sender.sendSwapFinishAck(seat, optReq.Tiles)
	if s.allPlayersSwapped() {
		s.executeSwap()
	}
	return nil
}

func (s *StateSwapTiles) allPlayersSwapped() bool {
	for _, st := range s.swapTiles {
		if st == nil {
			return false
		}
	}
	return true
}

func (s *StateSwapTiles) executeSwap() {
	swapType := rand.Int31n(3)
	switch swapType {
	case 1:
		s.swapClockwise()
	case 2:
		s.swapCross()
	default:
		s.swapCounterClockwise()
	}

	for _, st := range s.swapTiles {
		tiles := mahjong.Int32Tile(st.Tiles)
		s.game.play.GetPlayData(st.From).SwapOut(tiles)
		s.game.play.GetPlayData(st.To).SwapIn(tiles)
	}
	s.game.sender.sendSwapTilesResultAck(swapType, s.swapTiles)
	s.game.SetNextState(NewStateDingque)
}

func (s *StateSwapTiles) swapClockwise() {
	count := s.game.GetPlayerCount()
	for i := range count {
		s.swapTiles[i].To = (i - 1 + count) % count
	}
}

func (s *StateSwapTiles) swapCounterClockwise() {
	count := s.game.GetPlayerCount()
	for i := range count {
		s.swapTiles[i].To = (i + 1) % count
	}
}

func (s *StateSwapTiles) swapCross() {
	count := s.game.GetPlayerCount()
	for i := range count {
		s.swapTiles[i].To = (i + 2) % count
	}
}

func (s *StateSwapTiles) OnTimeout() {
	if s.game.MatchType == "fdtable" {
		return
	}
	for i := int32(0); i < s.game.GetPlayerCount(); i++ {
		if s.swapTiles[i] == nil {
			tiles := mahjong.TilesInt32(s.game.play.GetPlayData(i).GetSwapRecommend())
			s.game.sender.sendSwapFinishAck(i, tiles)
			s.swapTiles[i] = &pbsc.SCSwapTiles{
				From:  i,
				Tiles: tiles, // 随机换3张牌
			}
		}
	}
	s.executeSwap()
}
