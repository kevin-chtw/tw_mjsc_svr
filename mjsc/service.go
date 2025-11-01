package mjsc

import (
	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
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
		defaultRules: []int{10, 8, 1, 0},
		huCore:       mahjong.NewHuCore(14),
	}
	s.init()
	return s
}
func (s *service) init() {
	for color := mahjong.ColorCharacter; color <= mahjong.ColorDragon; color++ {
		pc := mahjong.PointCountByColor[color]
		for i := range pc {
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

func (s *service) GetFdRules() map[string]int32 {
	return nil
}

func (s *service) GetHuTypes(data *mahjong.HuData) []int32 {
	return newHuData(data).getHuTypes()
}

func (s *service) TotalMuti(types []int32, conf *mahjong.Rule) int64 {
	return totalMuti(types, conf)
}
