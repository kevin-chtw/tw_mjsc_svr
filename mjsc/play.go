package mjsc

import "github.com/kevin-chtw/tw_common/gamebase/mahjong"

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
	p.RegisterSelfCheck(newCheckerHu(p), newCheckerKon(p))
	p.RegisterWaitCheck(newCheckerPao(p), newCheckerZhiKon(p), newCheckerPon(p))
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

func (p *Play) discard(tile mahjong.Tile) bool {
	seat := p.GetCurSeat()
	queTile := p.getQueTile(seat)
	if queTile != mahjong.TileNull && tile.Color() != queTile.Color() {
		tile = queTile
	}
	return p.Play.Discard(tile)
}

func (p *Play) getQueTile(seat int32) mahjong.Tile {
	color := p.queColors[seat]
	tiles := p.GetPlayData(seat).GetHandTiles()
	for _, t := range tiles {
		if t.Color() == color {
			return t
		}
	}
	return mahjong.TileNull
}

func (p *Play) CheckHu(data *mahjong.HuData) mahjong.HuCoreType {
	if p.getQueTile(data.GetSeat()) != mahjong.TileNull {
		return mahjong.HU_NON
	}

	tiles, laiCount := data.CountLaiZi()
	htype := mahjong.Check7dui(tiles, laiCount)
	if htype != mahjong.HU_NON {
		return htype
	}
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
	types := make([]int32, 0)
	seat := p.GetCurSeat()
	if !p.HasOperate(seat) {
		if seat == p.GetBanker() {
			types = append(types, TianHu)
		} else {
			types = append(types, DiHu)
		}
		return types
	}

	types = append(types, ZiMo)
	if p.IsAfterKon() {
		types = append(types, KonKai)
	}
	if p.dealer.GetRestCount() == 0 {
		types = append(types, HaiDi)
	}
	return types
}

func (p *Play) paoHuTypes(_ int32) []int32 {
	types := make([]int32, 0)
	types = append(types, PaoHu)
	if p.IsAfterKon() {
		types = append(types, KonPao)
	}
	if p.dealer.GetRestCount() == 0 {
		types = append(types, HaiDiPao)
	}
	return types
}
