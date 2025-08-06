package lygc

type WinAckData struct {
	WinMode   int
	PaoSeat   int
	WinSeats  []int
	IsGrabKon bool
}

type Messager struct {
	game *Game
}

func NewMessager(game *Game) *Messager {
	return &Messager{
		game: game,
	}
}

func (m *Messager) SendDebugString(str string, seat int) {
	// 实现发送调试信息
}

func (m *Messager) SendGameStartAck() {
	// 实现发送游戏开始通知
}

func (m *Messager) SendPlaceAck() {
	// 实现发送摆牌通知
}

func (m *Messager) SendOpenDoorAck() {
	// 实现发送开门通知
}

func (m *Messager) SendOpenFanCiAck() {
	// 实现发送翻次通知
}

func (m *Messager) SendGenZhuangAck() {
	// 实现发送跟庄通知
}

func (m *Messager) SendPonAck(seat int) {
	// 实现发送碰牌通知
}

func (m *Messager) SendDiscardAck() {
	// 实现发送弃牌通知
}

func (m *Messager) SendDrawAck(showTiles map[int]int) {
	// 实现发送抽牌通知
}

func (m *Messager) SendWinAck(data WinAckData) {
	// 实现发送赢牌通知
}

func (m *Messager) SendTips(tipType int, seat int) {
	// 实现发送提示信息
}

func (m *Messager) SendHandTiles() {
	// 实现发送手牌信息
}

func (m *Messager) SendMahjongResult(isLiuJu bool, paoSeat, paoCiSeat int) {
	// 实现发送麻将结果
}

func (m *Messager) SendBeginAnimalAck() {
	// 实现发送开始动画通知
}

func (m *Messager) setUntrustOnGameEnd(seat int) {
	// 实现游戏结束时取消托管
}
