package mjsc

import "github.com/kevin-chtw/tw_common/gamebase/mahjong"

type checkerHu struct {
	play    *Play
	checker mahjong.CheckerSelf
}

func newCheckerHu(play *Play) mahjong.CheckerSelf {
	return &checkerHu{play: play, checker: mahjong.NewCheckerHu(play.Play)}
}

func (c *checkerHu) Check(opt *mahjong.Operates) {
	tile := c.play.getQueTile(c.play.GetCurSeat())
	if tile != mahjong.TileNull {
		return
	}
	c.checker.Check(opt)
}

// 杠检查器
type checkerKon struct {
	play *Play
}

func newCheckerKon(play *Play) mahjong.CheckerSelf {
	return &checkerKon{play: play}
}
func (c *checkerKon) Check(opt *mahjong.Operates) {
	if opt.IsMustHu || c.play.IsAfterPon() {
		return
	}
	seat := c.play.GetCurSeat()
	if c.play.GetPlayData(seat).CanSelfKonByQue(c.play.queColors[seat]) {
		opt.AddOperate(mahjong.OperateKon)
	}
}
