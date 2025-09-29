package mjsc

import (
	"maps"
	"slices"

	"github.com/kevin-chtw/tw_common/mahjong"
)

func init() {
	mahjong.Service = NewService()
}

type service struct {
	tiles        map[mahjong.Tile]int
	tilesFeng    map[mahjong.Tile]int
	defaultRules []int
	huCore       *mahjong.HuCore
}

func NewService() mahjong.IService {
	s := &service{
		tiles:        make(map[mahjong.Tile]int),
		tilesFeng:    make(map[mahjong.Tile]int),
		defaultRules: []int{10, 8},
		huCore:       mahjong.NewHuCore(14),
	}
	s.init()
	return s
}
func (s *service) init() {
	for color := mahjong.ColorCharacter; color <= mahjong.ColorDragon; color++ {
		pc := mahjong.PointCountByColor[color]
		for i := 0; i < pc; i++ {
			tile := mahjong.MakeTile(color, i)
			if color < mahjong.ColorWind {
				s.tiles[tile] = 4
			}
			s.tilesFeng[tile] = 4
		}
	}

}

func (s *service) GetAllTiles(conf *mahjong.Rule) map[mahjong.Tile]int {
	return s.tiles
}

func (s *service) GetHandCount() int {
	return 13
}

func (s *service) GetDefaultRules() []int {
	return s.defaultRules
}

func (s *service) CheckHu(data *mahjong.HuData, rule *mahjong.Rule) (*mahjong.HuResult, bool) {
	if !s.huCore.CheckBasicHu(data.TilesInHand, 0) {
		return nil, false
	}
	result := &mahjong.HuResult{
		HuTypes: make([]int32, len(data.ExtraHuTypes)),
	}
	copy(result.HuTypes, data.ExtraHuTypes)
	return result, true
}

func (s *service) CheckCall(data *mahjong.HuData, rule *mahjong.Rule) map[mahjong.Tile]map[mahjong.Tile]int64 {
	callData := make(map[mahjong.Tile]map[mahjong.Tile]int64)
	count := len(data.TilesInHand) % 3
	switch count {
	case 2:
		// 去重处理
		checkTiles := make([]mahjong.Tile, 0)
		tileSet := make(map[mahjong.Tile]bool)
		for _, tile := range data.TilesInHand {
			if !tileSet[tile] {
				tileSet[tile] = true
				checkTiles = append(checkTiles, tile)
			}
		}

		// 临时复制数据
		tempData := *data
		tempData.TilesInHand = slices.Clone(data.TilesInHand)

		for _, tile := range checkTiles {
			// 移除当前检查的牌
			tempData.TilesInHand = mahjong.RemoveElements(tempData.TilesInHand, tile, 1)

			// 检查叫牌
			fans := s.checkCallFan(&tempData, rule)
			if len(fans) > 0 {
				callData[tile] = make(map[mahjong.Tile]int64)
				maps.Copy(callData[tile], fans)
			}

			// 恢复牌
			tempData.TilesInHand = append(tempData.TilesInHand, tile)
		}
	case 1:
		// 直接检查叫牌
		fans := s.checkCallFan(data, rule)
		if len(fans) > 0 {
			callData[0] = fans
		}
	}

	return callData
}

func (s *service) checkCallFan(data *mahjong.HuData, rule *mahjong.Rule) map[mahjong.Tile]int64 {
	fans := make(map[mahjong.Tile]int64)
	testTiles := s.GetAllTiles(rule)
	originalTiles := make([]mahjong.Tile, len(data.TilesInHand))
	copy(originalTiles, data.TilesInHand)

	for tile := range testTiles {
		data.TilesInHand = append(data.TilesInHand, tile)
		if result, ok := s.CheckHu(data, rule); ok {
			fans[tile] = result.TotalMuti
		}
		data.TilesInHand = originalTiles
	}
	return fans
}
