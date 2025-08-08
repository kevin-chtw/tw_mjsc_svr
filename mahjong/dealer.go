package mahjong

import "math/rand"

// Dealer 麻将发牌器接口
type Dealer struct {
	game     *Game
	tileWall []ITileID
}

// NewDealer 创建新的发牌器
func NewDealer(game *Game) *Dealer {
	return &Dealer{
		game:     game,
		tileWall: make([]ITileID, 0),
	}
}

// GetGame 获取关联的Game对象
func (d *Dealer) GetGame() *Game {
	return d.game
}

func (d *Dealer) Initialize() {
	tiles := Service.GetAllTiles(d.game.config)
	// 预计算总牌数并一次性分配
	total := 0
	for _, count := range tiles {
		total += count
	}
	d.tileWall = make([]ITileID, total)

	// 填充并同时随机化牌墙
	i := 0
	for tile, count := range tiles {
		for j := 0; j < count; j++ {
			// 随机插入位置
			pos := rand.Intn(i + 1)
			if pos != i {
				d.tileWall[i] = d.tileWall[pos]
			}
			d.tileWall[pos] = tile
			i++
		}
	}
}

// DrawTile 抽牌
func (d *Dealer) DrawTile(drawSeat int32, playerData *PlayData) ITileID {
	// 抽牌逻辑
	return TileNull
}

// GetRestTileCount 获取剩余牌数
func (d *Dealer) GetRestCount() int {
	return len(d.tileWall)
}
