package lygc

import (
	"github.com/kevin-chtw/tw_game_svr/mahjong"
)

type CiData struct {
	Type EHuType
	Seat int
}

func NewCiData(typ EHuType, seat int) *CiData {
	return &CiData{
		Type: typ,
		Seat: seat,
	}
}

type PlayData struct {
	*mahjong.PlayData
	canTips map[int]bool    // 提示标记
	ciTiles map[int]*CiData // 次牌数据
}

func NewPlayData(seat int) *PlayData {
	return &PlayData{
		PlayData: mahjong.NewPlayData(seat),
		canTips:  make(map[int]bool),
		ciTiles:  make(map[int]*CiData),
	}
}

func (p *PlayData) CanSelfKon(cfg *mahjong.Config, ignoreTiles []int) bool {
	// 实现自杠检查逻辑
	return false
}

func (p *PlayData) CanSelfCi(cfg *mahjong.Config, ciTile int) bool {
	// 实现自次检查逻辑
	return false
}

func (p *PlayData) CanWaitCi(cfg *mahjong.Config, ciTile, outTile int) bool {
	// 实现等待次检查逻辑
	return false
}

func (p *PlayData) IsCanTips(tips int) bool {
	if can, ok := p.canTips[tips]; ok {
		return can
	}
	return false
}

func (p *PlayData) SetCanTips(tips int, canTips bool) {
	p.canTips[tips] = canTips
}

func (p *PlayData) GetCiTiles() []int {
	tiles := make([]int, 0, len(p.ciTiles))
	for tile := range p.ciTiles {
		tiles = append(tiles, tile)
	}
	return tiles
}

func (p *PlayData) GetCiType(ciTile int, seat *int) EHuType {
	if data, ok := p.ciTiles[ciTile]; ok {
		*seat = data.Seat
		return data.Type
	}
	return HuTypeNone
}

func (p *PlayData) isUsefulTile(tile int) bool {
	// 实现有用牌检查逻辑
	return false
}
