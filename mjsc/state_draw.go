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

		if !s.isCall(i) {
			s.tuiKon(i)
			s.chaJiao(i)
		}
	}

	s.game.sender.SendResult(true)
	s.WaitAni(s.game.OnGameOver)
}

func (s *StateDraw) tuiKon(seat int32) {
	if s.game.GetRule().GetValue(RuleTuiYu) == 0 {
		return
	}
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
	if s.game.GetRule().GetValue(RuleChaJiao) == 0 {
		return
	}
	multis := make([]int64, s.game.GetPlayerCount()) // 预分配数组
	for i := range s.game.GetPlayerCount() {
		if s.game.GetPlayer(i).IsOut() || i == seat {
			continue
		}

		maxMulti := s.maxMulti(seat)
		if maxMulti <= 0 {
			continue
		}
		multis[i] = maxMulti
		multis[seat] -= maxMulti
	}
	final := s.game.scorelator.CalcMulti(mahjong.SeatNull, mahjong.ScoreReasonChaJiao, multis)
	s.game.sender.SendScoreChangeAck(mahjong.ScoreReasonChaJiao, final, mahjong.TileNull, mahjong.SeatNull, nil)
}

func (s *StateDraw) isCall(seat int32) bool {
	callData := s.game.play.GetPlayData(seat).GetCallData()
	if len(callData) == 0 {
		return false
	}

	if s.game.GetRule().GetValue(RuleSiJBSJ) == 0 {
		return true
	}

	for t := range callData {
		if s.game.play.showCount(t) < 4 {
			return true
		}
	}
	return false
}

func (s *StateDraw) maxMulti(seat int32) int64 {
	if !s.isCall(seat) {
		return 0
	}
	maxMulti := int64(0)
	callData := s.game.play.GetPlayData(seat).GetCallData()
	for t, multi := range callData {
		if s.game.play.showCount(t) >= 4 {
			continue
		} else if multi > maxMulti {
			maxMulti = multi
		}
	}
	return maxMulti
}
