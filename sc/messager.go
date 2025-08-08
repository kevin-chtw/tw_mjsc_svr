package sc

import (
	"github.com/kevin-chtw/tw_game_svr/game"
	"github.com/kevin-chtw/tw_game_svr/mahjong"
	"github.com/kevin-chtw/tw_proto/scproto"
)

type WinAckData struct {
	WinMode   int
	PaoSeat   int
	WinSeats  []int
	IsGrabKon bool
}

type Messager struct {
	game *Game
	play *Play
}

func NewMessager(game *Game) *Messager {
	return &Messager{
		game: game,
		play: game.GetPlay().(*Play),
	}
}

func (m *Messager) sendGameStartAck() {
	startAck := &scproto.SCGameStartAck{
		Banker:    m.play.GetBanker(),
		TileCount: m.play.GetDealer().GetRestCount(),
		Scores:    m.play.GetCurScores(),
	}
	ack := &scproto.SCAck{Ack: &scproto.SCAck_ScGameStartAck{ScGameStartAck: startAck}}
	m.game.Send2Player(ack, game.SeatAll)
}

func (m *Messager) sendOpenDoorAck() {
	ack := &scproto.SCAck{Ack: &scproto.SCAck_ScOpenDoorAck{}}
	count := m.game.GetPlayerCount()
	for i := range count {
		openDoor := &scproto.SCOpenDoorAck{
			Seat:  i,
			Tiles: m.play.GetPlayData(i).GetHandTiles(),
		}
		ack.Ack.(*scproto.SCAck_ScOpenDoorAck).ScOpenDoorAck = openDoor
		m.game.Send2Player(openDoor, i)
	}
}

func (m *Messager) sendAnimationAck() {
	animationAck := &scproto.SCAnimationAck{
		Requestid: m.game.GetRequestID(game.SeatAll),
	}
	ack := &scproto.SCAck{Ack: &scproto.SCAck_ScAnimationAck{ScAnimationAck: animationAck}}
	m.game.Send2Player(ack, game.SeatAll)
}

func (m *Messager) sendRequestAck(seat int32, operates *mahjong.Operates) {
	requestAck := &scproto.SCRequestAck{
		Seat:        seat,
		RequestType: int32(operates.Value),
		Requestid:   m.game.GetRequestID(seat),
	}
	ack := &scproto.SCAck{Ack: &scproto.SCAck_ScRequestAck{ScRequestAck: requestAck}}
	m.game.Send2Player(ack, seat)
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

func (m *Messager) setUntrustOnGameEnd(seat int) {
	// 实现游戏结束时取消托管
}
