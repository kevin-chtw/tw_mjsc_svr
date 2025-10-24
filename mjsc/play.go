package mjsc

import "github.com/kevin-chtw/tw_common/mahjong"

type Play struct {
	*mahjong.Play
	dealer    *mahjong.Dealer
	queColors map[int32]mahjong.EColor
}

func NewPlay(game *Game) *Play {
	p := &Play{
		dealer:    mahjong.NewDealer(game.Game),
		queColors: make(map[int32]mahjong.EColor),
	}
	p.Play = mahjong.NewPlay(p, game.Game, p.dealer)
	p.PlayConf = &mahjong.PlayConf{}
	p.RegisterSelfCheck(mahjong.NewCheckerHu(p.Play), mahjong.NewCheckerTing(p.Play), mahjong.NewCheckerKon(p.Play))
	p.RegisterWaitCheck(mahjong.NewCheckerPao(p.Play), mahjong.NewCheckerPao(p.Play), mahjong.NewCheckerZhiKon(p.Play))
	return p
}

func (p *Play) queRecommand(seat int32) mahjong.EColor {
	tiles := p.GetPlayData(seat).GetHandTiles()
	colors := make(map[mahjong.EColor]int32)
	for _, tile := range tiles {
		colors[tile.Color()]++
	}

	bestColor := mahjong.ColorCharacter
	min := colors[mahjong.ColorCharacter]
	for c := mahjong.ColorCharacter + 1; c <= mahjong.ColorDot; c++ {
		if count, ok := colors[c]; !ok {
			min = 0
			bestColor = c
		} else if count < min {
			min = count
			bestColor = c
		}
	}
	return bestColor
}

func (p *Play) CheckHu(data *mahjong.HuData) bool {
	return data.CanHu()
}

func (p *Play) GetExtraHuTypes(playData *mahjong.PlayData, self bool) []int32 {
	if self || p.PlayConf.OnlyZimo {
		return p.selfHuTypes()
	} else {
		return p.paoHuTypes(playData.GetSeat())
	}
}

func (p *Play) selfHuTypes() []int32 {
	return []int32{HuTypeZiMo}
}

func (p *Play) paoHuTypes(seat int32) []int32 {
	return []int32{HuTypePingHu}
}
