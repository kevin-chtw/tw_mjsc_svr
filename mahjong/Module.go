package mahjong

// GameModule 麻将游戏模块基类
type GameModule struct {
	game *Game
}

// NewGameModule 创建新的游戏模块实例
func NewGameModule(game *Game) *GameModule {
	return &GameModule{
		game: game,
	}
}

// BrokerReqData 代理请求数据结构
type BrokerReqData struct {
	SerialsNo   int
	URL         string
	Extra       string
	HTTPHeader  string
	HTTPBody    string
}

// BrokerAckData 代理响应数据结构
type BrokerAckData struct {
	SerialsNo   int
	Extra       string
	HTTPHeader  string
	HTTPBody    string
}