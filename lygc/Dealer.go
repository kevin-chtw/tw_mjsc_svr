package lygc

import (
	"github.com/kevin-chtw/tw_game_svr/mahjong"
)

type ESDealType int

const (
	DealTypeAnKe       ESDealType = iota // 暗刻
	DealTypeAnGang                       // 暗杠
	DealTypeDoubleAnKe                   // 双暗刻
	DealTypeDuiZi                        // 对子
	DealTypePingHu                       // 平胡
)

type Dealer struct {
	*mahjong.Dealer
	ciTile             int
	ciAssociationCards []int
}

func NewDealer(game *mahjong.Game) *Dealer {
	return &Dealer{
		Dealer: mahjong.NewDealer(game),
	}
}

func (d *Dealer) Initialize() {
	d._InitCiTile()
}

func (d *Dealer) TestCtrl() bool {
	return d._CtrlDispatch()
}

func (d *Dealer) GetCiTile() int {
	return d.ciTile
}

func (d *Dealer) GetCiAssociationCards() []int {
	return d.ciAssociationCards
}

func (d *Dealer) GetGame() *mahjong.Game {
	if d.Dealer == nil {
		return nil
	}
	return d.Dealer.GetGame()
}

func (d *Dealer) _CtrlDispatch() bool {
	// 实现控制分发逻辑
	return false
}

type CiTilePoint int

const (
	CiTilePoint1_9 CiTilePoint = iota
	CiTilePoint2_8
	CiTilePoint3_7
	CiTilePointZi
)

func (d *Dealer) _InitCiTile() {
	// 实现初始化次牌逻辑
}

func (d *Dealer) _getCiTilePoint() CiTilePoint {
	// 实现获取次牌点逻辑
	return CiTilePoint1_9
}

func (d *Dealer) _checkFitCiPoint(tile int, ciPoint CiTilePoint) bool {
	// 实现检查适合的次牌点逻辑
	return false
}

func (d *Dealer) _getDuiZiCard(seat, count int) {
	// 实现获取对子牌逻辑
}

func (d *Dealer) _getShunCard(seat int) {
	// 实现获取顺子牌逻辑
}

func (d *Dealer) _getKeCard(seat, count int) {
	// 实现获取刻子牌逻辑
}

func (d *Dealer) _getGangCard(seat, count int) {
	// 实现获取杠牌逻辑
}

func (d *Dealer) _putCardIntoHand(seat, card, count int) {
	// 实现将牌放入手牌逻辑
}

func (d *Dealer) TilesMap() map[int]int {
	// 实现获取牌映射表逻辑
	return nil
}

func (d *Dealer) _takeCardsRandom(seat, count int) {
	// 实现随机取牌逻辑
}

func (d *Dealer) _checkRandomDealCard(typ ESDealType, callStep, seat int) bool {
	// 实现检查随机发牌逻辑
	return false
}

func (d *Dealer) _checkFanXing(tileCounts map[int]int, typ ESDealType) bool {
	// 实现检查番型逻辑
	return false
}

func (d *Dealer) _swapCards(seat, pos int) {
	// 实现交换牌逻辑
}

func (d *Dealer) _clearHandCards(seat int) {
	// 实现清空手牌逻辑
}

func (d *Dealer) _takeFanXingCards(seat int, typ ESDealType) {
	// 实现获取番型牌逻辑
}

func (d *Dealer) _takeCiAssociationCards(seat, count int) {
	// 实现获取次牌关联牌逻辑
}

func (d *Dealer) _checkCiCardPos() {
	// 实现检查次牌位置逻辑
}

func (d *Dealer) _DrawStrategy(typer *mahjong.PlayData) int {
	// 实现抽牌策略逻辑
	return 0
}

type NBDealer struct {
	*Dealer
	nbSeat int
}

func NewNBDealer(game *mahjong.Game, seat int) *NBDealer {
	return &NBDealer{
		Dealer: NewDealer(game),
		nbSeat: seat,
	}
}

func (d *NBDealer) Initialize() {
	// 实现新手发牌初始化逻辑
}

func (d *NBDealer) _baseHuNB(callStep, nNeedHua int) {
	// 实现基础新手胡牌逻辑
}

func (d *NBDealer) _getNBGangCiCard(seat, callStep int) {
	// 实现新手杠次牌逻辑
}

func (d *NBDealer) _getNBCiCardShun(seat int) {
	// 实现新手次牌顺子逻辑
}
