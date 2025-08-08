package mahjong

import "github.com/kevin-chtw/tw_game_svr/game"

type EHuResultMode int

const (
	HuResultNone EHuResultMode = iota
	HuResultPaoHu
	HuResultZiMo
)

// 玩家状态常量
const (
	PlayerStatusNone     int = iota // 无状态
	PlayerStatusPlaying             // 游戏中
	PlayerStatusWaiting             // 等待中
	PlayerStatusFinished            // 已完成
)

type Player struct {
	player      *game.Player
	game        *Game
	status      int // PlayerStatus
	huMode      EHuResultMode
	tax         int64
	score       int64 // 税前初始分
	isTrusted   bool
	isOffline   bool
	isOut       bool
	scoreChange int64 // 变动值(不含税)

}

func NewPlayer(game *Game, player *game.Player) *Player {
	return &Player{
		game:        game,
		player:      player,
		scoreChange: 0,
	}
}

func (p *Player) GetSeat() int32 {
	return p.player.Seat
}

func (p *Player) GetScore() int64 {
	return p.score
}

func (p *Player) GetUserID() string {
	return p.player.ID
}

func (p *Player) AddScoreChange(value int64) int64 {
	p.scoreChange += value
	return p.scoreChange
}

func (p *Player) GetIncremental(value int64) int64 {
	// 获取是否展示税后值
	return value
}

func (p *Player) GetScoreChange() int64 {
	return p.scoreChange
}

func (p *Player) GetScoreChangeWithTax() int64 {
	return p.scoreChange - p.tax
}

func (p *Player) GetCurrentScore() int64 {
	return p.score + p.scoreChange - p.tax
}

func (p *Player) SetTrusted(trusted bool) {
	p.isTrusted = trusted
}

func (p *Player) IsTrusted() bool {
	return p.isTrusted
}

func (p *Player) SetOffline(offline bool) {
	p.isOffline = offline
}

func (p *Player) IsOffline() bool {
	return p.isOffline
}

func (p *Player) SetOut(status int, mode EHuResultMode) {
	p.status = status
	p.huMode = mode
	p.isOut = true
}

func (p *Player) IsOut() bool {
	return p.isOut
}

func (p *Player) PayTax() {
	// 扣税逻辑
}

func (p *Player) RefundTax() {
	// 退税逻辑
}

func (p *Player) GetTax() int64 {
	return p.tax
}

func (p *Player) SyncGameResult() {
	// 同步游戏结果
}

func (p *Player) AddOperateEvent(chowCount, ponCount, konCount int, isCall bool) {
	// 添加操作事件统计
}

func (p *Player) AddHuEvent() {
	// 添加胡牌事件统计
}

func (p *Player) GetStatus() int {
	return p.status
}

func (p *Player) ResetScore(score int64) {
	p.score = score
	p.scoreChange = 0
	p.tax = 0
}

type GamePlayer struct {
	growValues map[uint32]int64
	callback   func(int32, uint32, int)
}

func NewGamePlayer() *GamePlayer {
	return &GamePlayer{
		growValues: make(map[uint32]int64),
	}
}

func (p *GamePlayer) SetValueByID(growID uint32, value int64) {
	p.growValues[growID] = value
}

func (p *GamePlayer) GetValueByID(growID uint32, notFoundValue int64) int64 {
	if val, ok := p.growValues[growID]; ok {
		return val
	}
	return notFoundValue
}

func (p *GamePlayer) GetHonorValue() int64 {
	// 获取荣誉值
	return 0
}

func (p *GamePlayer) SetCallback(callback func(int32, uint32, int)) {
	p.callback = callback
}

func (p *GamePlayer) OnPivotMsg(ackMsg interface{}, seatBind int, growIDBind uint32) bool {
	// 处理消息
	return true
}
