package mahjong

import "github.com/sirupsen/logrus"

type IPlay interface {
	Initialize()
	DoDeal()
	FetchSelfOperates() *Operates
	AddHistory(seat int32, tile ITileID, operate int, extra int)
}

type Play struct {
	game        *Game
	dealer      *Dealer
	currentSeat int32
	currentTile ITileID
	banker      int32
	history     []Action
}

func NewPlay(game *Game) *Play {
	return &Play{
		game:        game,
		dealer:      NewDealer(game),
		currentSeat: SeatNull,
		currentTile: TileNull,
		banker:      SeatNull,
	}
}

func (p *Play) Initialize() {
	lgd := p.getLastGameData()
	p.banker = lgd.banker
	p.currentSeat = p.banker
	p.dealer.Initialize()
	p.history = make([]Action, 0)
}

func (p *Play) getLastGameData() *LastGameData {
	lastGameData := p.game.GetLastGameData()
	if lastGameData == nil {
		return NewLastGameData(int(p.game.GetPlayerCount()))
	}

	lgd, ok := lastGameData.(*LastGameData)
	if !ok {
		logrus.Errorf("invalid last game data type: %T", lastGameData)
		return NewLastGameData(int(p.game.GetPlayerCount()))
	}
	return lgd
}

func (p *Play) DoDeal() {

}

func (p *Play) FetchSelfOperates() *Operates {
	// 获取当前玩家的操作
	return &Operates{}
}

func (p *Play) AddHistory(seat int32, tile ITileID, operate int, extra int) {
	action := Action{
		Seat:    seat,
		Tile:    tile,
		Operate: operate,
		Extra:   extra,
	}
	p.history = append(p.history, action)
}

func (p *Play) DoSwitchSeat(seat int32) {
	if seat == SeatNull {
		p.currentSeat = GetNextSeat(p.currentSeat, 1, int(p.game.GetPlayerCount()))
	} else {
		p.currentSeat = seat
	}
}

func (p *Play) GetCurrentSeat() int32 {
	return p.currentSeat
}

func (p *Play) GetCurrentTile() ITileID {
	return p.currentTile
}

func (p *Play) GetBanker() int32 {
	return p.banker
}
