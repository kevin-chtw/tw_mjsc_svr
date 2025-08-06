package mahjong

import (
	"container/list"
)

// Dealer 麻将发牌器接口
type Dealer struct {
	game      *Game
	tileWall  *list.List
	selfTiles map[ISeatID]*list.List
	isManual  bool
}

// NewDealer 创建新的发牌器
func NewDealer(game *Game) *Dealer {
	return &Dealer{
		game:      game,
		tileWall:  list.New(),
		selfTiles: make(map[ISeatID]*list.List),
	}
}

// GetGame 获取关联的Game对象
func (d *Dealer) GetGame() *Game {
	return d.game
}

// Initialize 初始化发牌器
func (d *Dealer) Initialize() {}

// DrawTile 抽牌
func (d *Dealer) DrawTile(drawSeat ISeatID, playerData *PlayData) ITileID {
	if d.isManual {
		return d.drawSelfTiles(drawSeat)
	}
	return d.drawTile(TileNull)
}

// GetTile 获取指定位置的牌
func (d *Dealer) GetTile(index int) ITileID {
	if index >= d.tileWall.Len() {
		return TileNull
	}
	e := d.tileWall.Front()
	for i := 0; i < index; i++ {
		e = e.Next()
	}
	return e.Value.(ITileID)
}

// GetRestTileCount 获取剩余牌数
func (d *Dealer) GetRestTileCount() int {
	return d.tileWall.Len()
}

// Exchange 交换牌
func (d *Dealer) Exchange(tiles []ITileID) {
	// 交换逻辑实现
}

// LoadManual 加载手动配置
func (d *Dealer) LoadManual(name string, initCardsFileNo, playerCount, bankerTileCount int) bool {
	d.isManual = true
	// 加载逻辑实现
	return true
}

// InitRandom 随机初始化
func (d *Dealer) InitRandom() {
	d.isManual = false
	// 随机初始化逻辑
}

// drawStrategy 抽牌策略接口
type drawStrategy interface {
	Draw(dealer *Dealer, typer *PlayData, configIndex int) ITileID
	DrawIndex(types []DrawTileType, configIndex int) int
	IsValid() bool
}

// DrawTileType 抽牌类型
type DrawTileType int

// DrawConfig 抽牌配置
type DrawConfig struct {
	DrawCount int
	Weights   []int
}

// baseDrawStrategy 基础抽牌策略
type baseDrawStrategy struct {
	game    *Game
	configs []DrawConfig
}

// NewDrawStrategy 创建新的抽牌策略
func NewDrawStrategy(game *Game) *baseDrawStrategy {
	return &baseDrawStrategy{
		game: game,
	}
}

// Init 初始化策略
func (s *baseDrawStrategy) Init(configs []DrawConfig) {
	s.configs = configs
}

// IsValid 检查策略是否有效
func (s *baseDrawStrategy) IsValid() bool {
	return len(s.configs) > 0
}

// Draw 执行抽牌
func (s *baseDrawStrategy) Draw(dealer *Dealer, typer *PlayData, configIndex int) ITileID {
	count := dealer.GetRestTileCount()
	if count == 0 {
		return TileNull
	}

	ci := s.fixConfigIndex(configIndex)
	drawCount := s.getDrawCount(ci)
	if count < drawCount {
		drawCount = count
	}
	if drawCount <= 1 {
		return dealer.GetTile(0)
	}

	types := make([]DrawTileType, drawCount)
	for i := 0; i < drawCount; i++ {
		tile := dealer.GetTile(i)
		types[i] = typer.CheckType(tile)
	}

	return dealer.GetTile(s.DrawIndex(types, configIndex))
}

// DrawIndex 计算抽牌索引
func (s *baseDrawStrategy) DrawIndex(types []DrawTileType, configIndex int) int {
	if len(types) < 2 || len(s.configs) == 0 {
		return 0
	}
	configIndex = s.fixConfigIndex(configIndex)
	weights := make([]int, len(types))
	for i, t := range types {
		if t < 0 || int(t) >= len(s.configs) {
			weights[i] = 0
		} else {
			weights[i] = s.configs[configIndex].Weights[int(t)]
		}
	}
	return RandomByRates(weights)
}

func (s *baseDrawStrategy) fixConfigIndex(configIndex int) int {
	if configIndex >= len(s.configs) {
		if len(s.configs) == 0 {
			return 0
		}
		return len(s.configs) - 1
	}
	return configIndex
}

func (s *baseDrawStrategy) getDrawCount(configIndex int) int {
	if len(s.configs) == 0 {
		return 0
	}
	return s.configs[s.fixConfigIndex(configIndex)].DrawCount
}

// 私有方法实现
func (d *Dealer) drawSelfTiles(drawSeat ISeatID) ITileID {
	// 私有牌堆抽牌实现
	return TileNull
}

func (d *Dealer) drawTile(need ITileID) ITileID {
	// 抽牌实现
	return TileNull
}

func (d *Dealer) drawStrategy(typer *PlayData) ITileID {
	// 抽牌策略实现
	return TileNull
}
