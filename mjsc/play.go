package mjsc

import "github.com/kevin-chtw/tw_common/mahjong"

type Play struct {
	*mahjong.Play
	dealer *mahjong.Dealer
}

func NewPlay(game *Game) *Play {
	p := &Play{
		dealer: mahjong.NewDealer(game.Game),
	}
	p.Play = mahjong.NewPlay(p, game.Game, p.dealer)
	p.PlayConf = &mahjong.PlayConf{}
	p.RegisterSelfCheck(mahjong.NewCheckerHu(p.Play), mahjong.NewCheckerTing(p.Play), mahjong.NewCheckerKon(p.Play))
	p.RegisterWaitCheck(mahjong.NewCheckerPao(p.Play), mahjong.NewCheckerPao(p.Play), mahjong.NewCheckerZhiKon(p.Play))
	return p
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
