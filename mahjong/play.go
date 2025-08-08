package mahjong

import "github.com/sirupsen/logrus"

type IPlay interface {
}

type Play struct {
	game        *Game
	dealer      *Dealer
	currentSeat int32
	currentTile int32
	banker      int32
	history     []Action
	playData    []*PlayData
}

func NewPlay(game *Game) *Play {
	return &Play{
		game:        game,
		dealer:      NewDealer(game),
		currentSeat: SeatNull,
		currentTile: TileNull,
		banker:      SeatNull,
		history:     make([]Action, 0),
		playData:    make([]*PlayData, game.GetPlayerCount()),
	}
}

func (p *Play) Initialize() {
	lgd := p.getLastGameData()
	p.banker = lgd.banker
	p.currentSeat = p.banker
	p.dealer.Initialize()
	p.history = make([]Action, 0)
}

func (p *Play) GetDealer() *Dealer {
	return p.dealer
}

func (p *Play) GetCurScores() []int64 {
	count := p.game.GetPlayerCount()
	scores := make([]int64, count)

	for i := range count {
		if player := p.game.GetPlayer(i); player != nil {
			scores[i] = player.GetCurScore()
		}
	}
	return scores
}

func (p *Play) Deal() {
	count := p.game.GetPlayerCount()
	for i := range count {
		p.playData[i].handTiles = p.dealer.Deal(Service.GetHandCount())
	}

	p.playData[p.banker].PutHandTile(p.dealer.DrawTile())
}

func (p *Play) GetPlayData(i int32) *PlayData {
	return p.playData[i]
}

func (p *Play) FetchSelfOperates() *Operates {
	// 获取当前玩家的操作
	return &Operates{}
}

func (p *Play) DoSwitchSeat(seat int32) {
	if seat == SeatNull {
		p.currentSeat = GetNextSeat(p.currentSeat, 1, int(p.game.GetPlayerCount()))
	} else {
		p.currentSeat = seat
	}
}

func (p *Play) GetCurSeat() int32 {
	return p.currentSeat
}

func (p *Play) GetCurTile() int32 {
	return p.currentTile
}

func (p *Play) GetBanker() int32 {
	return p.banker
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

func (p *Play) addHistory(seat int32, tile int32, operate int, extra int) {
	action := Action{
		Seat:    seat,
		Tile:    tile,
		Operate: operate,
		Extra:   extra,
	}
	p.history = append(p.history, action)
}
