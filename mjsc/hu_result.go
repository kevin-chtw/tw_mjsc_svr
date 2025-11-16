package mjsc

import (
	"slices"

	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/kevin-chtw/tw_proto/game/pbmj"
)

const (
	// huMode
	PaoHu      = 1 //点炮胡
	ZiMo       = 2 //自摸胡
	KonKai     = 3 //杠开
	KonPao     = 4 //杠炮
	QiangKonHu = 5 //抢杠胡
	HaiDi      = 6 //海底
	HaiDiPao   = 7 //海底炮
	TianHu     = 8 //天胡
	DiHu       = 9 //地胡

	// huType
	PingHu        = 20 //平胡
	PonPonHu      = 21 //碰碰胡
	QiDui         = 22 //七对
	QinYiSe       = 23 //清一色
	LongQiDui     = 24 //龙七对
	JinGouDiao    = 25 //金勾勾
	QinPon        = 26 //清碰
	QinQiDui      = 27 //清七对
	QinJinGouDiao = 28 //清金勾
	QinLongQiDui  = 29 //清龙七对

	// specialHuType
	JiaXinWu    = 100 //夹心五
	JueZhang    = 101 //绝张
	KaZhang     = 102 //卡张
	YiTiaoLong  = 103 //一条龙
	MenQing     = 104 //门清
	ZhongZhang  = 105 //中张
	JiangDui19  = 106 //幺九将对
	JiangDui258 = 107 //258将对
	BianZhang   = 108 //边张
)

var addMulti = map[int32]int64{
	JiaXinWu:    2,
	JueZhang:    2,
	KaZhang:     2,
	YiTiaoLong:  2,
	MenQing:     2,
	ZhongZhang:  2,
	JiangDui19:  8,
	JiangDui258: 8,
	BianZhang:   2,
}

var multis = map[int32]int64{
	// 胡牌模式
	KonKai:   2,
	KonPao:   2,
	HaiDi:    2,
	HaiDiPao: 2,
	TianHu:   16,
	DiHu:     16,
	// 基础番型
	PonPonHu:      2,
	QiDui:         4,
	QinYiSe:       4,
	LongQiDui:     8,
	JinGouDiao:    8,
	QinPon:        8,
	QinQiDui:      16,
	QinJinGouDiao: 16,
	QinLongQiDui:  32,
}

// 定义番型检查配置
type huTypeConfig struct {
	checkFunc func(*HuData) bool
	huType    int32
	exclude   []int32
}

// 全局番型配置
var huConfigs = []huTypeConfig{
	// 基础番型
	{(*HuData).isLongQiDui, LongQiDui, []int32{QiDui}},
	{(*HuData).isQinYiSe, QinYiSe, []int32{PingHu}},
	{(*HuData).isJinGouDiao, JinGouDiao, []int32{PonPonHu}},

	// 组合番型
	{(*HuData).isQinPon, QinPon, []int32{QinYiSe, PonPonHu}},
	{(*HuData).isQinQiDui, QinQiDui, []int32{QinYiSe, QiDui}},
	{(*HuData).isQingLongQiDui, QinLongQiDui, []int32{QinYiSe, LongQiDui}},
	{(*HuData).isQinJinGouDiao, QinJinGouDiao, []int32{QinYiSe, JinGouDiao}},

	// 其他番型
	{(*HuData).isYiTiaoLong, YiTiaoLong, nil},
	{(*HuData).isMenQing, MenQing, nil},
	{(*HuData).isZhongZhang, ZhongZhang, nil},
	{(*HuData).isJiangDui19, JiangDui19, nil},
	{(*HuData).isJueZhang, JueZhang, nil},
	{(*HuData).isJiangDui258, JiangDui258, nil},
	{(*HuData).isBianZhang, BianZhang, nil},
	{(*HuData).isKaZhang, KaZhang, nil},
	{(*HuData).isJiaWuXing, JiaXinWu, []int32{KaZhang}},
}

func totalMuti(result *pbmj.MJHuData, conf *mahjong.Rule) int64 {
	totalMuti := int64(1 << result.Gen)
	for k, v := range multis {
		if slices.Contains(result.HuTypes, k) {
			totalMuti *= v
		}
	}
	if slices.Contains(result.HuTypes, ZiMo) {
		if conf.GetValue(RuleZiMoJiaDi) == 1 {
			totalMuti += 1
		} else {
			totalMuti *= 2
		}
	}
	for k, v := range addMulti {
		if slices.Contains(result.HuTypes, k) {
			totalMuti += v
		}
	}
	return totalMuti
}

type HuData struct {
	*mahjong.HuData
	CheckFunc func(*HuData) bool // 自定义检查函数
}

func newHuData(data *mahjong.HuData) *HuData {
	return &HuData{
		HuData: data,
	}
}

// Checkfunc 调用自定义检查函数
func (h *HuData) Checkfunc() bool {
	if h.CheckFunc != nil {
		return h.CheckFunc(h)
	}
	return false
}
func (h *HuData) calcGen() int32 {
	// 获取玩家数据
	playData := h.PlayData
	genCount := int32(len(playData.GetKonGroups()))

	// 2. 计算碰牌与手牌组成4张的牌数
	ponGroups := playData.GetPonGroups()
	handTiles := playData.GetHandTiles()

	tileCount := make(map[mahjong.Tile]int32)
	for _, tile := range handTiles {
		tileCount[tile]++
	}

	for _, pon := range ponGroups {
		tileCount[pon.Tile] += 3
	}

	for _, count := range tileCount {
		if count == 4 {
			genCount++
		}
	}
	return genCount
}

func (h *HuData) getHuTypes() []int32 {
	types := slices.Clone(h.ExtraHuTypes)
	if slices.Contains(types, TianHu) || slices.Contains(types, DiHu) {
		return types
	}
	switch h.HuCoreType {
	case mahjong.HU_7DUI:
		types = append(types, QiDui)
	case mahjong.HU_PON:
		types = append(types, PonPonHu)
	default:
		types = append(types, PingHu)
	}
	for _, config := range huConfigs {
		types = h.check(types, config)
	}
	return types
}
func (h *HuData) check(types []int32, cfg huTypeConfig) []int32 {
	if !cfg.checkFunc(h) {
		return types
	}
	newTypes := make([]int32, 0, len(types))
	for _, t := range types {
		if !slices.Contains(cfg.exclude, t) {
			newTypes = append(newTypes, t)
		}
	}
	return append(newTypes, cfg.huType)
}

func (h *HuData) isQinYiSe() bool {
	tiles := h.Tiles
	if len(tiles) == 0 {
		return false
	}

	firstColor := tiles[0].Color()
	for _, tile := range tiles {
		if tile.Color() != firstColor {
			return false
		}
	}
	playData := h.PlayData
	for _, g := range playData.GetPonGroups() {
		if g.Tile.Color() != firstColor {
			return false
		}
	}
	for _, g := range playData.GetKonGroups() {
		if g.Tile.Color() != firstColor {
			return false
		}
	}
	return true
}

func (h *HuData) isLongQiDui() bool {
	if h.HuCoreType != mahjong.HU_7DUI {
		return false
	}

	tileMap := make(map[mahjong.Tile]int)
	for _, tile := range h.Tiles {
		tileMap[tile]++
	}
	for _, count := range tileMap {
		if count == 4 {
			return true
		}
	}
	return false
}

func (h *HuData) isJinGouDiao() bool {
	return len(h.Tiles) == 2
}

func (h *HuData) isQinPon() bool {
	return h.isQinYiSe() && h.HuCoreType == mahjong.HU_PON
}

func (h *HuData) isQinJinGouDiao() bool {
	return h.isQinYiSe() && h.isJinGouDiao()
}

func (h *HuData) isQinQiDui() bool {
	return h.isQinYiSe() && h.HuCoreType == mahjong.HU_7DUI
}

func (h *HuData) isQingLongQiDui() bool {
	return h.isQinYiSe() && h.isLongQiDui()
}

func (h *HuData) isYiTiaoLong() bool {
	if h.Play.GetRule().GetValue(RuleYiTiaoLong) == 0 {
		return false
	}

	tiles := h.Tiles
	if len(tiles) < 9 {
		return false
	}
	// 按花色分组
	colorGroups := make(map[mahjong.EColor][]mahjong.Tile)
	for _, tile := range tiles {
		if tile.IsSuit() {
			colorGroups[tile.Color()] = append(colorGroups[tile.Color()], tile)
		}
	}

	// 检查每个花色组是否包含1-9
	for _, group := range colorGroups {
		if len(group) < 9 {
			continue
		}

		points := make(map[int]bool)
		for _, tile := range group {
			points[tile.Point()] = true
		}

		hasAll := true
		for i := range 9 {
			if !points[i] {
				hasAll = false
				break
			}
		}
		if hasAll {
			return true
		}
	}
	return false
}

func (h *HuData) isMenQing() bool {
	if h.Play.GetRule().GetValue(RuleMQZZ) == 0 {
		return false
	}

	playData := h.PlayData
	if len(playData.GetPonGroups()) > 0 {
		return false
	}

	konGroups := playData.GetKonGroups()
	for _, kon := range konGroups {
		if kon.Type != mahjong.KonTypeAn { // 非暗杠（即明杠）
			return false
		}
	}
	return true
}

func (h *HuData) isZhongZhang() bool {
	if h.Play.GetRule().GetValue(RuleMQZZ) == 0 {
		return false
	}

	isZhongTile := func(p int) bool { return p > 0 && p < 8 }
	return h.checkAllTiles(isZhongTile)
}

func (h *HuData) isJueZhang() bool {
	if h.Play.GetRule().GetValue(RuleJueZhang) == 0 {
		return false
	}
	play := h.Play.PlayImp.(*Play)
	count := play.showCount(h.CurTile)
	if h.Self {
		count += 1
	}
	return count >= 4
}

func (h *HuData) isJiangDui19() bool {
	if h.Play.GetRule().GetValue(RuleJiangDui19) == 0 {
		return false
	}
	is19 := func(p int) bool { return p == 0 || p == 8 }
	return h.checkAllTiles(is19)
}

func (h *HuData) isJiangDui258() bool {
	if h.Play.GetRule().GetValue(RuleJiangDui258) == 0 {
		return false
	}
	// 允许的点数集合（麻将内部点数范围0-8，这里只允许1、4、7）
	is258 := func(p int) bool { return p == 1 || p == 4 || p == 7 }
	return h.checkAllTiles(is258)
}

// checkAllTiles 统一校验：手牌 + 碰 + 杠 的点数是否满足谓词
func (h *HuData) checkAllTiles(pred func(int) bool) bool {
	for _, t := range h.Tiles {
		if !pred(t.Point()) {
			return false
		}
	}
	pd := h.PlayData
	for _, g := range pd.GetPonGroups() {
		if !pred(g.Tile.Point()) {
			return false
		}
	}
	for _, g := range pd.GetKonGroups() {
		if !pred(g.Tile.Point()) {
			return false
		}
	}
	return true
}

func (h *HuData) isKaZhang() bool {
	if h.Play.GetRule().GetValue(RuleKaBianZhang) == 0 {
		return false
	}
	if len(h.PlayData.GetCallData()) > 1 { //卡张仅一个听口
		return false
	}
	waitTile := h.CurTile
	point := waitTile.Point()

	if point <= 0 || point >= 8 { // 在0-8范围内，1-7需要检查相邻牌
		return false
	}

	return h.CheckShun(waitTile, point-1, point+1)
}

func (h *HuData) isBianZhang() bool {
	if h.Play.GetRule().GetValue(RuleKaBianZhang) == 0 {
		return false
	}
	waitTile := h.CurTile
	point := waitTile.Point()

	switch point {
	case 2:
		return h.CheckShun(waitTile, point-1, point-2) && !h.CheckShun(waitTile, point+1, point+2)
	case 6:
		return !h.CheckShun(waitTile, point-1, point-2) && h.CheckShun(waitTile, point+1, point+2)
	default:
		return false
	}
}

func (h *HuData) isJiaWuXing() bool {
	if h.Play.GetRule().GetValue(RuleJiaXinWu) == 0 {
		return false
	}
	if h.isKaZhang() && h.CurTile.Point() == 4 {
		return true
	}
	return false
}
