package mjsc

import "github.com/kevin-chtw/tw_common/gamebase/mahjong"

type checkerHu struct {
	play    *Play
	checker mahjong.CheckerSelf
}

func newCheckerHu(play *Play) mahjong.CheckerSelf {
	return &checkerHu{play: play, checker: mahjong.NewCheckerHu(play.Play)}
}

func (c *checkerHu) Check(opt *mahjong.Operates, tips []int) []int {
	tile := c.play.getQueTile(c.play.GetCurSeat())
	if tile != mahjong.TileNull {
		return tips
	}
	return c.checker.Check(opt, tips)
}

// 杠检查器
type checkerKon struct {
	play *Play
}

func newCheckerKon(play *Play) mahjong.CheckerSelf {
	return &checkerKon{play: play}
}
func (c *checkerKon) Check(opt *mahjong.Operates, tips []int) []int {
	if opt.IsMustHu {
		return tips
	}
	seat := c.play.GetCurSeat()
	if c.play.GetPlayData(seat).CanSelfKonByQue(c.play.queColors[seat]) {
		opt.AddOperate(mahjong.OperateKon)
	}
	return tips
}
