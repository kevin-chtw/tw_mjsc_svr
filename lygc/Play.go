package lygc

import "github.com/kevin-chtw/tw_game_svr/mahjong"

type Play struct {
	*mahjong.Play
	isGenZhuang    bool
	touchTileTimes map[int]int
}

func NewPlay(game *Game) *Play {
	return &Play{
		Play:           mahjong.NewPlay(game.Game),
		touchTileTimes: make(map[int]int),
	}
}

func (p *Play) Initialize() {
	// 实现初始化逻辑
}

func (p *Play) DoCheckGangMultiple() []int64 {
	// 实现检查杠倍数逻辑
	return nil
}

func (p *Play) DoCheckLiuJuMultiple() []int64 {
	// 实现检查流局倍数逻辑
	return nil
}

func (p *Play) DoCheckGenZhuangMultiple() []int64 {
	// 实现检查跟庄倍数逻辑
	return nil
}

func (p *Play) DoCheckMiMultiple(data *LastHandData, huType EHuType) []int64 {
	// 实现检查密牌倍数逻辑
	return nil
}

func (p *Play) CheckCurGenZhuang() bool {
	// 实现检查当前跟庄逻辑
	return false
}

func (p *Play) IsGenZhuang() bool {
	return p.isGenZhuang
}

func (p *Play) GetCurSeatTouchTimes() int {
	// 实现获取当前座位摸牌次数逻辑
	return 0
}

func (p *Play) SetCurSeatTouchTimes(times int) {
	// 实现设置当前座位摸牌次数逻辑
}

func (p *Play) ziMoExtraFan(seat int) []int {
	// 实现自摸额外番数逻辑
	return nil
}

func (p *Play) paoExtraFan(seat int, extraFan []int) bool {
	// 实现跑胡额外番数逻辑
	return false
}

func (p *Play) qiangGangExtraFan(seat int) []int {
	// 实现抢杠额外番数逻辑
	return nil
}
