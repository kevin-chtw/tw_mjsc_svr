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
	BianZhang   = 108 //卡张
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
	{isLongQiDui, LongQiDui, []int32{QiDui}},
	{isQinYiSe, QinYiSe, []int32{PingHu}},
	{isJinGouDiao, JinGouDiao, []int32{PonPonHu}},

	// 组合番型
	{isQinPon, QinPon, []int32{QinYiSe, PonPonHu}},
	{isQinQiDui, QinQiDui, []int32{QinYiSe, QiDui}},
	{isQingLongQiDui, QinLongQiDui, []int32{QinYiSe, LongQiDui}},
	{isQinJinGouDiao, QinJinGouDiao, []int32{QinYiSe, JinGouDiao}},

	// 其他番型
	{isYiTiaoLong, YiTiaoLong, nil},
	{isMenQing, MenQing, nil},
	{isZhongZhang, ZhongZhang, nil},
	{isJiangDui19, JiangDui19, nil},
	{isJueZhang, JueZhang, nil},
	{isJiangDui258, JiangDui258, nil},
	{isBianZhang, BianZhang, nil},
	{isKaZhang, KaZhang, nil},
	{isJiaWuXing, JiaXinWu, []int32{KaZhang}},
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
}

func newHuData(data *mahjong.HuData) *HuData {
	return &HuData{
		HuData: data,
	}
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
	playData := huData.PlayData
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

func isMenQing(huData *HuData) bool {
	playData := huData.PlayData

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

func isZhongZhang(huData *HuData) bool {
	playData := huData.PlayData

	// 1. 检查手牌
	for _, tile := range playData.GetHandTiles() {
		if !isZhongTile(tile) {
			return false
		}
	}

	// 2. 检查碰牌
	for _, pon := range playData.GetPonGroups() {
		if !isZhongTile(pon.Tile) {
			return false
		}
	}

	// 3. 检查杠牌
	for _, kon := range playData.GetKonGroups() {
		if !isZhongTile(kon.Tile) {
			return false
		}
	}

	return true
}

// 辅助函数：判断单张牌是否是中张
func isZhongTile(tile mahjong.Tile) bool {
	if tile.IsSuit() {
		point := tile.Point()
		return point >= 1 && point <= 7 // 在0-8范围内，1-7算中张
	}
	return false // 字牌不算中张
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
	tile := huData.CurTile
	totalCount := 0
	for i := range huData.Play.GetPlayerCount() {
		playData := huData.Play.GetPlayData(i)
		for _, pon := range playData.GetPonGroups() {
			if pon.Tile == tile {
				totalCount += 3
			}
		}
		for _, t := range playData.GetOutTiles() {
			if t == tile {
				totalCount += 1
			}
		}
	}
	if huData.Self {
		totalCount += 1
	}
	return totalCount >= 4
}

func isJiangDui258(huData *HuData) bool {
	tiles := huData.Tiles
	for _, tile := range tiles {
		point := tile.Point()
		if point != 1 && point != 4 && point != 7 {
			return false
		}
	}
	return true
}

func isKaZhang(huData *HuData) bool {
	if len(huData.PlayData.GetCallData()) > 1 { //卡张仅一个听口
		return false
	}
	waitTile := huData.CurTile
	point := waitTile.Point()

	if point <= 0 || point >= 8 { // 在0-8范围内，1-7需要检查相邻牌
		return false
	}

	return huData.CheckShun(waitTile, point-1, point+1)
}

func isBianZhang(huData *HuData) bool {
	waitTile := huData.CurTile
	point := waitTile.Point()

	switch point {
	case 2:
		return huData.CheckShun(waitTile, point-1, point-2) && !huData.CheckShun(waitTile, point+1, point+2)
	case 6:
		return !huData.CheckShun(waitTile, point-1, point-2) && huData.CheckShun(waitTile, point+1, point+2)
	default:
		return false
	}
}

func isJiaWuXing(huData *HuData) bool {
	if isKaZhang(huData) && huData.CurTile.Point() == 4 {
		return true
	}
	return false
}
