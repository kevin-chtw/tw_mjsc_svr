package mjsc

import (
	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
)

type StateDraw struct {
	*State
}

func NewStateDraw(game mahjong.IGame, args ...any) mahjong.IState {
	return &StateDraw{
		State: NewState(game),
	}
}

func (s *StateDraw) OnEnter() {
	if s.game.GetRestCount() == 1 {
		s.game.sender.SendResult(false)
		s.WaitAni(s.game.OnGameOver)
		return
	}

	tile := s.game.play.Draw()
	if tile == mahjong.TileNull {
		s.liuJu()
		return
	}
	s.game.sender.SendDrawAck(tile)
	s.game.SetNextState(NewStateDiscard)
}

func (s *StateDraw) liuJu() {
	for i := range s.game.GetPlayerCount() {
		if s.game.GetPlayer(i).IsOut() {
			continue
		}
		callData := s.game.play.GetPlayData(i).GetCallData()
		if len(callData) == 0 {
			s.tuiKon(i)
			s.chaJiao(i)
		}
	}

	s.game.sender.SendResult(true)
	s.WaitAni(s.game.OnGameOver)
}

func (s *StateDraw) tuiKon(seat int32) {
	scoreNodes := s.game.scorelator.GetKonScores(seat)
	for _, sn := range scoreNodes {
		scores := make([]int64, len(sn.Scores))
		konSeatScore := int64(0)
		for i, v := range sn.Scores {
			if int32(i) != seat {
				if s.game.MatchType == "fdtable" || !s.game.GetPlayer(int32(i)).IsOut() {
					konSeatScore += v
					scores[i] = -v
				}
			}
		}
		scores[seat] = konSeatScore
		final := s.game.scorelator.CalcScores(mahjong.SeatNull, mahjong.ScoreReasonTuiKon, scores)
		s.game.sender.SendScoreChangeAck(mahjong.ScoreReasonTuiKon, final, mahjong.TileNull, mahjong.SeatNull, nil)
	}
}

func (s *StateDraw) chaJiao(seat int32) {
	multis := make([]int64, s.game.GetPlayerCount()) // 预分配数组
	for i := range s.game.GetPlayerCount() {
		if s.game.GetPlayer(i).IsOut() || i == seat {
			continue
		}

		callData := s.game.play.GetPlayData(i).GetCallData()
		if len(callData) == 0 {
			continue
		}
		maxMulti := int64(0)
		for _, v := range callData {
			if v > maxMulti {
				maxMulti = v
			}
		}
		multis[i] = maxMulti
		multis[seat] -= maxMulti
	}
	final := s.game.scorelator.CalcMulti(mahjong.SeatNull, mahjong.ScoreReasonChaJiao, multis)
	s.game.sender.SendScoreChangeAck(mahjong.ScoreReasonChaJiao, final, mahjong.TileNull, mahjong.SeatNull, nil)
}
