package mjsc

import (
	"slices"

	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
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
	KaBianZhang = 102 //卡边张
	YiTiaoLong  = 103 //一条龙
	MenQing     = 104 //门清
	ZhongZhang  = 105 //中张
	JiangDui19  = 106 //幺九将对
	JiangDui258 = 107 //258将对
)

var addMulti = map[int32]int64{
	JiaXinWu:    2,
	JueZhang:    2,
	KaBianZhang: 2,
	YiTiaoLong:  2,
	MenQing:     2,
	ZhongZhang:  2,
	JiangDui19:  8,
	JiangDui258: 8,
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
	{isLongQiDui, LongQiDui, []int32{QiDui}},
	{isQinYiSe, QinYiSe, []int32{PingHu}},
	{isJinGouDiao, JinGouDiao, []int32{PonPonHu}},

	// 组合番型
	{isQinPon, QinPon, []int32{QinYiSe, PonPonHu}},
	{isQinQiDui, QinQiDui, []int32{QinYiSe, QiDui}},
	{isQingLongQiDui, QinLongQiDui, []int32{QinYiSe, LongQiDui}},
	{isQinJinGouDiao, QinJinGouDiao, []int32{QinYiSe, JinGouDiao}},

	// 其他番型
	{isJiaWuXing, JiaXinWu, []int32{KaBianZhang}}, // 夹心五与卡边张互斥
	{isYiTiaoLong, YiTiaoLong, nil},
	{isMenQing, MenQing, nil},
	{isZhongZhang, ZhongZhang, nil},
	{isJiangDui19, JiangDui19, []int32{PonPonHu}}, // 幺九将对与碰碰胡互斥
	{isJueZhang, JueZhang, nil},
	{isKaBianZhang, KaBianZhang, nil}, // 卡边张与夹心五互斥
	{isJiangDui258, JiangDui258, nil}, // 将对258可与七对叠加
}

func totalMuti(huTypes []int32, conf *mahjong.Rule) int64 {
	totalMuti := int64(1)
	for k, v := range multis {
		if slices.Contains(huTypes, k) {
			totalMuti *= v
		}
	}
	if slices.Contains(huTypes, ZiMo) {
		if conf.GetValue(RuleZiMoJiaDi) == 1 {
			totalMuti += 1
		} else {
			totalMuti *= 2
		}
	}
	for k, v := range addMulti {
		if slices.Contains(huTypes, k) {
			totalMuti += v
		}
	}
	return totalMuti
}

type HuData struct {
	*mahjong.HuData
	huCache map[int32]bool // 番型计算结果缓存
}

func newHuData(data *mahjong.HuData) *HuData {
	return &HuData{
		HuData:  data,
		huCache: make(map[int32]bool),
	}
}

func (h *HuData) getHuTypes() []int32 {
	types := []int32{PingHu}
	for _, config := range huConfigs {
		h.check(types, config)
	}
	return types
}

func (h *HuData) check(types []int32, cfg huTypeConfig) []int32 {
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

	if !cfg.checkFunc(h) {
		h.huCache[cfg.huType] = false
		return types
	}
	newTypes := make([]int32, 0, len(types))
	for _, t := range types {
		if !slices.Contains(cfg.exclude, t) {
			newTypes = append(newTypes, t)
		}
	}
	h.huCache[cfg.huType] = true
	return append(newTypes, cfg.huType)
}

func isQinYiSe(huData *HuData) bool {
	tiles := huData.Tiles
	if len(tiles) == 0 {
		return false
	}

	firstColor := tiles[0].Color()
	for _, tile := range tiles {
		if tile.Color() != firstColor {
			return false
		}
	}
	return true
}

func isLongQiDui(huData *HuData) bool {
	if huData.HuCoreType != mahjong.HU_7DUI {
		return false
	}

	tileMap := make(map[mahjong.Tile]int)
	for _, tile := range huData.Tiles {
		tileMap[tile]++
	}
	for _, count := range tileMap {
		if count == 4 {
			return true
		}
	}
	return false
}

func isJinGouDiao(huData *HuData) bool {
	return len(huData.Tiles) == 2
}

func isQinPon(huData *HuData) bool {
	return isQinYiSe(huData) && huData.HuCoreType == mahjong.HU_PON
}

func isQinJinGouDiao(huData *HuData) bool {
	return isQinYiSe(huData) && isJinGouDiao(huData)
}

func isQinQiDui(huData *HuData) bool {
	return isQinYiSe(huData) && huData.HuCoreType == mahjong.HU_7DUI
}

func isQingLongQiDui(huData *HuData) bool {
	return isQinYiSe(huData) && isLongQiDui(huData)
}

func isJiaWuXing(huData *HuData) bool {
	waitTile := huData.GetCurTile()
	if !waitTile.IsSuit() || waitTile.Point() != 5 {
		return false // 不是5筒/5万/5条
	}

	tiles := huData.Tiles
	tileMap := make(map[mahjong.Tile]int)
	for _, tile := range tiles {
		tileMap[tile]++
	}

	color := waitTile.Color()
	return tileMap[mahjong.MakeTile(color, 4)] > 0 && tileMap[mahjong.MakeTile(color, 6)] > 0
}

func isYiTiaoLong(huData *HuData) bool {
	tiles := huData.Tiles
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
		for i := 1; i <= 9; i++ {
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

func isMenQing(huData *HuData) bool {
	return false
}

func isZhongZhang(huData *HuData) bool {
	return isJiaWuXing(huData)
}

func isJiangDui19(huData *HuData) bool {
	tiles := huData.Tiles
	for _, tile := range tiles {
		if tile.IsHonor() {
			continue // 字牌算幺九
		}
		if point := tile.Point(); point != 1 && point != 9 {
			return false
		}
	}
	return true
}

func isJueZhang(huData *HuData) bool {
	return false
}

func isKaBianZhang(huData *HuData) bool {
	waitTile := huData.GetCurTile()
	if !waitTile.IsSuit() {
		return false
	}

	point := waitTile.Point()
	if point == 3 || point == 7 {
		// 边张情况
		return true
	}

	// 检查卡张情况
	tiles := huData.Tiles
	tileMap := make(map[mahjong.Tile]int)
	for _, tile := range tiles {
		tileMap[tile]++
	}

	color := waitTile.Color()
	if point > 1 && point < 9 {
		// 检查是否有相邻的牌
		return tileMap[mahjong.MakeTile(color, point-1)] > 0 &&
			tileMap[mahjong.MakeTile(color, point+1)] > 0
	}
	return false
}

func isJiangDui258(huData *HuData) bool {
	tiles := huData.Tiles
	for _, tile := range tiles {
		if tile.IsHonor() {
			return false // 不能有字牌
		}
		point := tile.Point()
		if point != 2 && point != 5 && point != 8 {
			return false
		}
	}
	return true
}
