package mahjong

type TileType int

const (
	TileUseful   TileType = iota // 有效牌
	TileSpecial                  // 特殊牌
	TileRelated                  // 相关牌
	TileUseless                  // 无效牌
)

type PlayDataEx struct {
	*PlayData
	*GameModule
	gangTiles     []int
	removeHuTypes map[int]struct{}
}

func NewPlayDataEx(game *Game, seat int) *PlayDataEx {
	return &PlayDataEx{
		PlayData:   NewPlayData(seat),
		GameModule: NewGameModule(game),
		gangTiles:    make([]int, 0),
		removeHuTypes: make(map[int]struct{}),
	}
}

func (p *PlayDataEx) CanPon(tile int, cantOnlyLaiAfterPon bool) bool {
	// 实现碰牌判断逻辑
	return false
}

func (p *PlayDataEx) CanChow(tile int) bool {
	// 实现吃牌判断逻辑
	return false
}

func (p *PlayDataEx) CanZhiKon(tile int, cfg *Config) bool {
	// 实现直杠判断逻辑
	return false
}

func (p *PlayDataEx) CanSelfKon(cfg *Config, ignoreTiles []int) bool {
	// 实现自杠判断逻辑
	return false
}

func (p *PlayDataEx) CanDiscard(tile int) bool {
	// 实现弃牌判断逻辑
	return false
}

func (p *PlayDataEx) CanKon(tile int, konType KonType) bool {
	// 实现杠牌判断逻辑
	return false
}

func (p *PlayDataEx) CanChowWithLeft(leftTile, curTile int) bool {
	// 实现带左边牌的吃牌判断逻辑
	return false
}

func (p *PlayDataEx) IsTianDiHu() bool {
	// 实现天地胡判断逻辑
	return false
}

func (p *PlayDataEx) IsCurrentDrawTile(tile int) bool {
	// 实现当前抽牌判断逻辑
	return false
}

func (p *PlayDataEx) FindAnyInHand(tiles map[int]struct{}) int {
	// 实现查找手牌中任意匹配牌的逻辑
	return 0
}

func (p *PlayDataEx) CloseTianDiHu() {
	// 关闭天地胡状态
}

func (p *PlayDataEx) PickTianHuTile() {
	// 实现天胡牌选择逻辑
}

func (p *PlayDataEx) ClearDraw() {
	// 清空抽牌状态
}

func (p *PlayDataEx) SetRemoveHutypes(types map[int]struct{}) {
	p.removeHuTypes = make(map[int]struct{})
	for typ := range types {
		p.removeHuTypes[typ] = struct{}{}
	}
}

func (p *PlayDataEx) GetHandStyle() EHandStyle {
	// 实现获取手牌风格逻辑
	return HandNone
}

func (p *PlayDataEx) MakeHuPlayData(extraTile, removeTile int, removeCount int) *HuPlayData {
	// 转换手牌类型为ITileID
	handTiles := make([]ITileID, len(p.GetHandTiles()))
	for i, tile := range p.GetHandTiles() {
		handTiles[i] = ITileID(tile)
	}
	
	data := &HuPlayData{
		TilesInHand:      handTiles,
		TilesForChowLeft: make([]ITileID, 0),
		TilesForPon:     make([]ITileID, 0),
		TilesForKon:     make([]ITileID, 0),
		PaoTile:         ITileID(extraTile),
		RemoveHuType:    p.removeHuTypes,
	}
	return data
}

func (p *PlayDataEx) GetGangTiles() []int {
	return p.gangTiles
}

func (p *PlayDataEx) IsAllLai() bool {
	// 实现全赖判断逻辑
	return false
}

func (p *PlayDataEx) canKonAfterCall(tile int, konType KonType, cfg *Config) bool {
	// 实现调用后杠牌判断逻辑
	return false
}

func (p *PlayDataEx) CheckType(tile int) int {
	// 实现牌类型检查逻辑
	if p.isUsefulTile(tile) {
		return int(TileUseful)
	}
	if p.isSpeacilTile(tile) {
		return int(TileSpecial)
	}
	if p.isRelateTile(tile) {
		return int(TileRelated)
	}
	return int(TileUseless)
}

func (p *PlayDataEx) UpdateMinTing(tile int) {
	// 实现最小听牌更新逻辑
}

func (p *PlayDataEx) isUsefulTile(tile int) bool {
	// 实现有用牌判断逻辑
	return false
}

func (p *PlayDataEx) isSpeacilTile(tile int) bool {
	// 实现特殊牌判断逻辑
	return false
}

func (p *PlayDataEx) isRelateTile(tile int) bool {
	// 实现相关牌判断逻辑
	return false
}