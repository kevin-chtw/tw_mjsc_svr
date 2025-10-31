package mjsc

import "github.com/kevin-chtw/tw_common/gamebase/mahjong"

const (
	// 基础番型 (按倍数从小到大)
	PingHu        = iota //四副顺子/刻子 + 一对将 0 (1倍)
	PonPonHu             //四副刻子+一对将 1 (2倍)
	QiDui                //7对独立牌 2 (4倍)
	QinYiSe              //全为同一花色 3 (4倍)
	LongQiDui            //七对子中含四张相同牌 4 (8倍)
	JinGouDiao           //碰碰胡+单吊 5 (8倍)
	QinPon               //清一色+碰碰胡 6 (8倍)
	QinQiDui             //清一色+七对 7 (16倍)
	QinJinGouDiao        //清一色+碰碰胡+单吊 8 (16倍)
	QingLongQiDui        //清一色+龙七对 9 (32倍)

	// 特殊番型 (按倍数从小到大)
	JiaWuXing        //夹心五(46胡5) 10 (2倍)
	JueZhang         //胡绝张 11 (2倍)
	KaBianZhang      //卡/边张(37或13卡2) 12 (2倍)
	YiTiaoLong       //同一花色的123456789 13 (4倍)
	MenQinZhongZhang //门清+2-8的牌 14 (4倍)
	YaoJiuJiangDui   //刻子/顺子全带幺九牌 15 (8倍)
	JiangDui258      //刻子和将对全由258组成 16 (8倍)
	TianHu           //庄家14张直接胡牌 17 (16倍)
	DiHu             //闲家自摸首轮胡牌 18 (16倍)
)

var multiples = map[int32]int64{
	// 基础番型
	PingHu:        1,  // 平胡
	PonPonHu:      2,  // 碰碰胡
	QiDui:         4,  // 七对
	QinYiSe:       4,  // 清一色
	LongQiDui:     8,  // 龙七对
	JinGouDiao:    8,  // 金勾勾
	QinPon:        8,  // 清碰
	QinQiDui:      16, // 清七对
	QinJinGouDiao: 16, // 清金勾
	QingLongQiDui: 32, // 清龙七对

	// 特殊番型
	JiaWuXing:        2,  // 夹心五
	JueZhang:         2,  // 胡绝张
	KaBianZhang:      2,  // 卡边张
	YiTiaoLong:       4,  // 一条龙
	MenQinZhongZhang: 4,  // 门清中张
	YaoJiuJiangDui:   8,  // 幺九将对
	JiangDui258:      8,  // 将对258
	TianHu:           16, // 天胡
	DiHu:             16, // 地胡
}

// 定义番型检查配置
type huTypeConfig struct {
	checkFunc func(*mahjong.HuData) bool
	huType    int32
	exclude   []int32
}

// 全局番型配置
var huConfigs = []huTypeConfig{
	// 基础番型
	{isPonPonHu, PonPonHu, []int32{PingHu}},
	{isQiDui, QiDui, []int32{PingHu}},
	{isLongQiDui, LongQiDui, []int32{PingHu}},
	{isQinYiSe, QinYiSe, []int32{PingHu}},
	{isJinGouDiao, JinGouDiao, []int32{PingHu}},

	// 组合番型
	{isQinPon, QinPon, []int32{QinYiSe, PonPonHu}},
	{isQinQiDui, QinQiDui, []int32{QinYiSe, QiDui}},
	{isQingLongQiDui, QingLongQiDui, []int32{QinYiSe, LongQiDui}},
	{isQinJinGouDiao, QinJinGouDiao, []int32{QinYiSe, JinGouDiao}},

	// 其他番型
	{isJiaWuXing, JiaWuXing, []int32{KaBianZhang}}, // 夹心五与卡边张互斥
	{isYiTiaoLong, YiTiaoLong, nil},
	{isMenQinZhongZhang, MenQinZhongZhang, nil},
	{isYaoJiuJiangDui, YaoJiuJiangDui, []int32{PonPonHu}}, // 幺九将对与碰碰胡互斥
	{isTianHu, TianHu, []int32{PingHu, PonPonHu, QiDui, QinYiSe, LongQiDui, JinGouDiao, YaoJiuJiangDui, JiangDui258, QinPon, QinQiDui, QinJinGouDiao, QingLongQiDui}}, // 天胡不与其他番型叠加
	{isDiHu, DiHu, []int32{PingHu, PonPonHu, QiDui, QinYiSe, LongQiDui, JinGouDiao, YaoJiuJiangDui, JiangDui258, QinPon, QinQiDui, QinJinGouDiao, QingLongQiDui}},     // 地胡不与其他番型叠加
	{isJueZhang, JueZhang, nil},
	{isKaBianZhang, KaBianZhang, []int32{JiaWuXing}}, // 卡边张与夹心五互斥
	{isJiangDui258, JiangDui258, nil},                // 将对258可与七对叠加
}

func totalMuti(huTypes []int32) int64 {
	totalMuti := int64(1)
	for _, huType := range huTypes {
		if multiple, ok := multiples[huType]; ok {
			totalMuti *= multiple
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

func (h *HuData) checkAndAppend(types []int32, checkFunc func(*mahjong.HuData) bool, huType int32, excludeTypes ...int32) []int32 {
	if checkFunc(h.HuData) {
		for _, excludeType := range excludeTypes {
			for i, t := range types {
				if t == excludeType {
					types = append(types[:i], types[i+1:]...)
					break
				}
			}
		}
		h.huCache[huType] = true
		return append(types, huType)
	}
	h.huCache[huType] = false
	return types
}

func (h *HuData) getHuTypes() []int32 {
	types := []int32{PingHu}
	for _, config := range huConfigs {
		if config.exclude != nil {
			types = h.checkAndAppend(types, config.checkFunc, config.huType, config.exclude...)
		} else {
			types = h.checkAndAppend(types, config.checkFunc, config.huType)
		}
	}
	return types
}

func isPonPonHu(huData *mahjong.HuData) bool {
	tiles := huData.Tiles
	tileMap := make(map[mahjong.Tile]int)
	for _, tile := range tiles {
		tileMap[tile]++
	}

	pengCount := 0
	jiangCount := 0
	for _, count := range tileMap {
		switch count {
		case 3:
			pengCount++
		case 2:
			jiangCount++
		}
	}

	return pengCount == 4 && jiangCount == 1
}

func isQiDui(huData *mahjong.HuData) bool {
	// 七对: 7对独立牌
	tiles := huData.Tiles
	if len(tiles) != 14 {
		return false
	}

	tileMap := make(map[mahjong.Tile]int)
	for _, tile := range tiles {
		tileMap[tile]++
	}

	for _, count := range tileMap {
		if count != 2 {
			return false
		}
	}
	return true
}

func isQinYiSe(huData *mahjong.HuData) bool {
	// 清一色: 全为同一花色
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

func isLongQiDui(huData *mahjong.HuData) bool {
	// 龙七对: 七对子中包含至少1个四张相同的牌
	tiles := huData.Tiles
	if len(tiles) != 14 {
		return false
	}

	tileMap := make(map[mahjong.Tile]int)
	for _, tile := range tiles {
		tileMap[tile]++
	}

	hasFour := false
	for _, count := range tileMap {
		if count == 4 {
			hasFour = true
		} else if count != 2 {
			return false
		}
	}
	return hasFour
}

func isJinGouDiao(huData *mahjong.HuData) bool {
	// 金勾勾: 碰碰胡+单吊
	if !isPonPonHu(huData) {
		return false
	}

	// 检查是否单吊
	tiles := huData.Tiles
	tileMap := make(map[mahjong.Tile]int)
	for _, tile := range tiles {
		tileMap[tile]++
	}

	jiangCount := 0
	for _, count := range tileMap {
		if count == 2 {
			jiangCount++
		}
	}
	return jiangCount == 1
}

func isQinPon(huData *mahjong.HuData) bool {
	// 清碰: 清一色+碰碰胡
	return isQinYiSe(huData) && isPonPonHu(huData)
}

func isQinJinGouDiao(huData *mahjong.HuData) bool {
	// 清金勾: 清一色+金勾勾
	return isQinYiSe(huData) && isJinGouDiao(huData)
}

func isQinQiDui(huData *mahjong.HuData) bool {
	// 清七对: 清一色+七对
	return isQinYiSe(huData) && isQiDui(huData)
}

func isQingLongQiDui(huData *mahjong.HuData) bool {
	// 清龙七对: 清一色+龙七对
	return isQinYiSe(huData) && isLongQiDui(huData)
}

func isJiaWuXing(huData *mahjong.HuData) bool {
	// 夹五星: 胡牌胡5筒、5万、5条的夹，比如46胡5
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
	// 检查是否有4和6
	return tileMap[mahjong.MakeTile(color, 4)] > 0 && tileMap[mahjong.MakeTile(color, 6)] > 0
}

func isYiTiaoLong(huData *mahjong.HuData) bool {
	// 一条龙: 手牌/胡牌能组成同一花色的123456789
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

func isMenQinZhongZhang(huData *mahjong.HuData) bool {
	// 门清中张: 门清+2-8的牌，即门清+断幺
	// 检查门清(没有吃、碰、明杠)
	// 由于 mahjong.HuData 没有直接提供 Groups() 方法
	// 这里暂时返回 false，需要根据实际业务逻辑实现
	return false
}

func isYaoJiuJiangDui(huData *mahjong.HuData) bool {
	// 幺九将对: 刻子、顺子全带幺九牌
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

func isTianHu(huData *mahjong.HuData) bool {
	// 天胡: 庄家14张牌直接满足胡牌条件
	// 需要根据 PlayData 判断是否是庄家
	//return huData.PlayData.Play.IsZhuang() && huData.PlayData.Play.IsFirstRound()
	return false
}

func isDiHu(huData *mahjong.HuData) bool {
	// 地胡: 闲家在没有碰、杠的干扰下，自摸胡牌
	// 需要根据 PlayData 判断是否是闲家、首轮和自摸
	// return !huData.PlayData.Play.IsZhuang() &&
	// 	huData.PlayData.Play.IsFirstRound() &&
	// 	huData.PlayData.Play.IsZiMo() &&
	// 	!huData.PlayData.Play.HasPeng() &&
	// 	!huData.PlayData.Play.HasGang()
	return false
}

func isJueZhang(huData *mahjong.HuData) bool {
	// 绝张: 胡的牌只有1张的情况下(其他三张在牌池中或者被碰已现)
	// waitTile := huData.GetCurTile()
	// 需要从 PlayData 获取已出现的牌数
	// 这里暂时返回 false，需要根据实际业务逻辑实现
	return false
}

func isKaBianZhang(huData *mahjong.HuData) bool {
	// 卡边张: 37边张胡，或者卡中间张，如13卡2胡
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

func isJiangDui258(huData *mahjong.HuData) bool {
	// 将对258: 刻子和将对全是由258组成
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
