package mahjong

var Service IService

type IService interface {
	GetAllTiles(conf *Config) map[ITileID]int
}
