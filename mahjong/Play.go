package mahjong

type Play struct {
	game        *Game
	currentSeat ISeatID
	currentTile ITileID
	banker      ISeatID
	history     []Action
	logger      Logger // 添加日志接口
}

func NewPlay(game *Game) *Play {
	return &Play{
		game:        game,
		currentSeat: SeatNull,
		currentTile: TileNull,
		banker:      SeatNull,
		logger:      game.GetLogger(), // 从game获取logger
	}
}

func (p *Play) Initialize() {
	// 初始化游戏玩法
	p.currentSeat = SeatNull
	p.currentTile = TileNull
	p.history = make([]Action, 0)
}

func (p *Play) DoDeal() {
	//发牌
}

func (p *Play) FetchSelfOperates() *Operates {
	// 获取当前玩家的操作
	return &Operates{}
}

func (p *Play) AddHistory(seat ISeatID, tile ITileID, operate int, extra int) {
	action := Action{
		Seat:    seat,
		Tile:    tile,
		Operate: operate,
		Extra:   extra,
	}
	p.history = append(p.history, action)
}

func (p *Play) DumpHistory() {
	// 打印历史记录
	for i, action := range p.history {
		p.logger.Printf("[History %d] Seat:%d Tile:%d Operate:%d Extra:%v\n",
			i, action.Seat, action.Tile, action.Operate, action.Extra)
	}
}

func (p *Play) DoSwitchSeat(seat ISeatID) {
	if seat == SeatNull {
		// 自动计算下一个座位
		seatCount := 4 // 默认4人麻将
		p.currentSeat = GetNextSeat(p.currentSeat, 1, seatCount)
	} else {
		p.currentSeat = seat
	}
}

func (p *Play) GetCurrentSeat() ISeatID {
	return p.currentSeat
}

func (p *Play) GetCurrentTile() ITileID {
	return p.currentTile
}

func (p *Play) GetBanker() ISeatID {
	return p.banker
}

type Action struct {
	Seat    ISeatID
	Tile    ITileID
	Operate int
	Extra   int
}
