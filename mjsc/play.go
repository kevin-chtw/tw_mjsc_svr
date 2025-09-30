package mjsc

import "github.com/kevin-chtw/tw_common/mahjong"

type Play struct {
	*mahjong.Play
}

func NewPlay(game *Game) *Play {
	p := &Play{
		Play: mahjong.NewPlay(game.Game),
	}
	p.ExtraHuTypes = p
	p.PlayConf = &mahjong.PlayConf{}
	p.RegisterSelfCheck(&mahjong.CheckerHu{}, &mahjong.CheckerTing{}, &mahjong.CheckerKon{})
	p.RegisterWaitCheck(&mahjong.CheckerPao{}, &mahjong.CheckerPon{}, &mahjong.CheckerZhiKon{})
	return p
}

func (p *Play) SelfExtraFans() []int32 {
	return []int32{}
}

func (p *Play) PaoExtraFans() []int32 {
	return []int32{}
}
