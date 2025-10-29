package mjsc

import "github.com/kevin-chtw/tw_common/gamebase/mahjong"

type checkerPao struct {
	play    *Play
	checker mahjong.CheckerWait
}

func newCheckerPao(play *Play) mahjong.CheckerWait {
	return &checkerPao{play: play, checker: mahjong.NewCheckerPao(play.Play)}
}
func (c *checkerPao) Check(seat int32, opt *mahjong.Operates, tips []int) []int {
	tile := c.play.GetCurTile()
	if tile.Color() == c.play.getQueTile(seat).Color() {
		return tips
	}
	return c.checker.Check(seat, opt, tips)
}

type checkerPon struct {
	play    *Play
	checker mahjong.CheckerWait
}

func newCheckerPon(play *Play) mahjong.CheckerWait {
	return &checkerPon{play: play, checker: mahjong.NewCheckerPon(play.Play)}
}
func (c *checkerPon) Check(seat int32, opt *mahjong.Operates, tips []int) []int {
	tile := c.play.GetCurTile()
	if tile.Color() == c.play.getQueTile(seat).Color() {
		return tips
	}
	return c.checker.Check(seat, opt, tips)
}

type checkerZhiKon struct {
	play    *Play
	checker mahjong.CheckerWait
}

func newCheckerZhiKon(play *Play) mahjong.CheckerWait {
	return &checkerZhiKon{play: play, checker: mahjong.NewCheckerZhiKon(play.Play)}
}
func (c *checkerZhiKon) Check(seat int32, opt *mahjong.Operates, tips []int) []int {
	tile := c.play.GetCurTile()
	if tile.Color() == c.play.getQueTile(seat).Color() {
		return tips
	}
	return c.checker.Check(seat, opt, tips)
}
