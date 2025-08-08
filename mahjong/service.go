package mahjong

var Service IService

type IService interface {
	GetAllTiles(conf *Rule) map[int32]int
	GetHandCount() int
	GetDefaultRules() []int
}
