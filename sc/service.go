package sc

import "github.com/kevin-chtw/tw_game_svr/mahjong"

var HuCore *mahjong.HuCore

func init() {
	mahjong.Service = NewService()
	HuCore = mahjong.NewHuCore(14)
}

type service struct {
	tiles     map[mahjong.ITileID]int
	tilesFeng map[mahjong.ITileID]int
}

func NewService() mahjong.IService {
	s := &service{
		tiles: make(map[mahjong.ITileID]int),
	}
	s.init()
	return s
}
func (s *service) init() {
	for color := mahjong.ColorCharacter; color <= mahjong.ColorDragon; color++ {
		pc := mahjong.PointCountByColor[color]
		for i := 0; i < pc; i++ {
			tile := mahjong.MakeTile(color, i, 0)
			if color < mahjong.ColorWind {
				s.tiles[tile] = 4
			}
			s.tilesFeng[tile] = 4
		}
	}
}

func (s *service) GetAllTiles(conf *mahjong.Config) map[mahjong.ITileID]int {
	return s.tiles
}
