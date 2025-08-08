package mahjong

type Group struct {
	Tile  int
	From  int
	Extra int
}

type KonGroup struct {
	Tile          int
	From          int
	Type          KonType
	HandPassBuKon bool
	Extra         int
}

type ChowGroup struct {
	ChowTile int
	From     int
	LeftTile int
}

type PassHuType map[int]int

type PlayData struct {
	callDataMap      map[int]map[int]int
	currentDrawTiles map[int]struct{}
	call             bool
	tianTing         bool
	handTiles        []int32
	outTiles         []int32
	tianDiHuState    bool
	passPon          map[int]struct{}
	passHu           PassHuType
	qiHuFanLimitTip  bool
	chowGroups       []ChowGroup
	ponGroups        []Group
	konGroups        []KonGroup
	everPonCount     int
	everKonCount     int
	everChiCount     int
	minTingValue     int
	drawConfig       int
	drawRate         int
}

func NewPlayData(seat int) *PlayData {
	return &PlayData{
		callDataMap:      make(map[int]map[int]int),
		currentDrawTiles: make(map[int]struct{}),
		handTiles:        make([]int32, 0),
		outTiles:         make([]int32, 0),
		passPon:          make(map[int]struct{}),
		passHu:           make(PassHuType),
		chowGroups:       make([]ChowGroup, 0),
		ponGroups:        make([]Group, 0),
		konGroups:        make([]KonGroup, 0),
		minTingValue:     17,
	}
}

func (p *PlayData) MutableCallDataMap() map[int]map[int]int {
	return p.callDataMap
}

func (p *PlayData) GetCallDataMap() map[int]map[int]int {
	return p.callDataMap
}

func (p *PlayData) Draw(tile int) {
	p.currentDrawTiles[tile] = struct{}{}
}

func (p *PlayData) Discard(tile int) {
	delete(p.currentDrawTiles, tile)
}

func (p *PlayData) SetCall(tile int, tianTing bool) {
	p.call = true
	p.tianTing = tianTing
}

func (p *PlayData) PutHandTile(tile int32) {
	p.handTiles = append(p.handTiles, tile)
}

func (p *PlayData) RemoveHandTile(tile int, count int) {
	// 实现移除多张手牌逻辑
}

func (p *PlayData) PutOutTile(tile int32) {
	p.outTiles = append(p.outTiles, tile)
}

func (p *PlayData) RemoveOutTile() {
	if len(p.outTiles) > 0 {
		p.outTiles = p.outTiles[:len(p.outTiles)-1]
	}
}

func (p *PlayData) HasTile(tile int32) bool {
	for _, t := range p.handTiles {
		if t == tile {
			return true
		}
	}
	return false
}

func (p *PlayData) GetHandTiles() []int32 {
	return p.handTiles
}

func (p *PlayData) GetOutTiles() []int32 {
	return p.outTiles
}

func (p *PlayData) CloseTianDiHu() {
	p.tianDiHuState = false
}

func (p *PlayData) TianDiHuState() bool {
	return p.tianDiHuState
}

func (p *PlayData) IsPassHuTile(tile, fan int) bool {
	if f, ok := p.passHu[tile]; ok {
		return f == fan
	}
	return false
}

func (p *PlayData) IsPassHuTileWithoutFan(tile int) bool {
	_, ok := p.passHu[tile]
	return ok
}

func (p *PlayData) IsPassPonTile(tile int) bool {
	_, ok := p.passPon[tile]
	return ok
}

func (p *PlayData) ClearPass() {
	p.passPon = make(map[int]struct{})
	p.passHu = make(PassHuType)
}

func (p *PlayData) PassPon(tile int) {
	p.passPon[tile] = struct{}{}
}

func (p *PlayData) PassHu(tile, fan int) {
	p.passHu[tile] = fan
}

func (p *PlayData) SetBanQiHuFanTip(flag bool) {
	p.qiHuFanLimitTip = flag
}

func (p *PlayData) IsBanQiHuFanTip() bool {
	return p.qiHuFanLimitTip
}

func (p *PlayData) Chow(curTile, tile, from int) int {
	group := ChowGroup{
		ChowTile: curTile,
		From:     from,
		LeftTile: tile,
	}
	p.chowGroups = append(p.chowGroups, group)
	return curTile
}

func (p *PlayData) GetChowGroups() []ChowGroup {
	return p.chowGroups
}

func (p *PlayData) Pon(tile, from int) int {
	group := Group{
		Tile: tile,
		From: from,
	}
	p.ponGroups = append(p.ponGroups, group)
	return tile
}

func (p *PlayData) HasPon(tile int) bool {
	for _, group := range p.ponGroups {
		if group.Tile == tile {
			return true
		}
	}
	return false
}

func (p *PlayData) KonS(tile int, konType KonType, from int, handPassBuKon, buKonAfterPon bool) int {
	group := KonGroup{
		Tile:          tile,
		Type:          konType,
		From:          from,
		HandPassBuKon: handPassBuKon,
	}
	p.konGroups = append(p.konGroups, group)
	return tile
}

func (p *PlayData) HasKon(tile int) bool {
	for _, group := range p.konGroups {
		if group.Tile == tile {
			return true
		}
	}
	return false
}

func (p *PlayData) PushPon(group Group) {
	p.ponGroups = append(p.ponGroups, group)
}

func (p *PlayData) PushKon(group KonGroup) {
	p.konGroups = append(p.konGroups, group)
}

func (p *PlayData) GetKon(tile int) *KonGroup {
	for _, group := range p.konGroups {
		if group.Tile == tile {
			return &group
		}
	}
	return nil
}

func (p *PlayData) GetPon(tile int) *Group {
	for _, group := range p.ponGroups {
		if group.Tile == tile {
			return &group
		}
	}
	return nil
}

func (p *PlayData) RemovePon(tile int) Group {
	for i, group := range p.ponGroups {
		if group.Tile == tile {
			p.ponGroups = append(p.ponGroups[:i], p.ponGroups[i+1:]...)
			return group
		}
	}
	return Group{}
}

func (p *PlayData) RemoveKon(tile int) KonGroup {
	for i, group := range p.konGroups {
		if group.Tile == tile {
			p.konGroups = append(p.konGroups[:i], p.konGroups[i+1:]...)
			return group
		}
	}
	return KonGroup{}
}

func (p *PlayData) RevertKon(tile int) {
	// 实现杠牌回退逻辑
}

func (p *PlayData) GetPonGroups() []Group {
	return p.ponGroups
}

func (p *PlayData) GetKonGroups() []KonGroup {
	return p.konGroups
}

func (p *PlayData) GetExchangeRecommend() []int {
	// 实现交换推荐逻辑
	return nil
}

func (p *PlayData) CanExchangeOut(tiles []int) bool {
	// 实现能否交换出牌逻辑
	return false
}

func (p *PlayData) ExchangeOut(outs []int) {
	// 实现交换出牌逻辑
}

func (p *PlayData) ExchangeIn(ines []int) {
	// 实现交换进牌逻辑
}

func (p *PlayData) Exchange(outs, ines []int) {
	p.ExchangeOut(outs)
	p.ExchangeIn(ines)
}

func (p *PlayData) IncEverPonCount() {
	p.everPonCount++
}

func (p *PlayData) IncEverKonCount() {
	p.everKonCount++
}

func (p *PlayData) IncEverChiCount() {
	p.everChiCount++
}

func (p *PlayData) GetEverPonCount() int {
	return p.everPonCount
}

func (p *PlayData) GetEverKonCount() int {
	return p.everKonCount
}

func (p *PlayData) GetEverChiCount() int {
	return p.everChiCount
}

func (p *PlayData) GetMinTing() int {
	return p.minTingValue
}

func (p *PlayData) SetDrawConfig(drawConfig, drawRate int) {
	p.drawConfig = drawConfig
	p.drawRate = drawRate
}

func (p *PlayData) GetDrawConfig() int {
	return p.drawConfig
}

func (p *PlayData) GetDrawRate() int {
	return p.drawRate
}
