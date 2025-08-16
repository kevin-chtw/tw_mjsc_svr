package sc

import (
	"github.com/kevin-chtw/tw_game_svr/game"
	"github.com/kevin-chtw/tw_game_svr/mahjong"
	"github.com/kevin-chtw/tw_proto/scproto"
)

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
	ack := &scproto.SCAck{ScGameStartAck: startAck}
	m.game.Send2Player(ack, game.SeatAll)
}

func (m *Messager) sendOpenDoorAck() {
	count := m.game.GetPlayerCount()
	for i := range count {
		openDoor := &scproto.SCOpenDoorAck{
			Seat:  i,
			Tiles: m.play.GetPlayData(i).GetHandTiles(),
		}
		ack := &scproto.SCAck{ScOpenDoorAck: openDoor}
		m.game.Send2Player(ack, i)
	}
}

func (m *Messager) sendAnimationAck() {
	animationAck := &scproto.SCAnimationAck{
		Requestid: m.game.GetRequestID(game.SeatAll),
	}
	ack := &scproto.SCAck{ScAnimationAck: animationAck}
	m.game.Send2Player(ack, game.SeatAll)
}

func (m *Messager) sendRequestAck(seat int32, operates *mahjong.Operates) {
	requestAck := &scproto.SCRequestAck{
		Seat:        seat,
		RequestType: int32(operates.Value),
		Requestid:   m.game.GetRequestID(seat),
	}
	ack := &scproto.SCAck{ScRequestAck: requestAck}
	m.game.Send2Player(ack, seat)
}

func (m *Messager) sendDiscardAck() {
	discardAck := &scproto.SCDiscardAck{
		Seat: m.play.GetCurSeat(),
		Tile: m.play.GetCurTile(),
	}
	ack := &scproto.SCAck{ScDiscardAck: discardAck}
	m.game.Send2Player(ack, game.SeatAll)
}

func (m *Messager) sendPonAck(seat int32) {
	ponAck := &scproto.SCPonAck{
		Seat: seat,
		From: m.play.GetCurSeat(),
		Tile: m.play.GetCurTile(),
	}
	ack := &scproto.SCAck{ScPonAck: ponAck}
	m.game.Send2Player(ack, game.SeatAll)
}

func (m *Messager) sendKonAck(seat, tile int32, konType mahjong.KonType) {
	konAck := &scproto.SCKonAck{
		Seat:    seat,
		From:    m.play.GetCurSeat(),
		Tile:    tile,
		KonType: int32(konType),
	}
	ack := &scproto.SCAck{ScKonAck: konAck}
	m.game.Send2Player(ack, game.SeatAll)
}

func (m *Messager) sendHuAck(huSeats []int32, paoSeat int32) {
	huAck := &scproto.SCHuAck{
		PaoSeat: paoSeat,
		Tile:    m.play.GetCurTile(),
		HuData:  make([]*scproto.SCHuData, len(huSeats)),
	}
	for i := range huSeats {
		huAck.HuData[i] = &scproto.SCHuData{
			Seat:    huSeats[i],
			HuTypes: m.play.GetHuResult(huSeats[i]).HuTypes,
		}
	}
	ack := &scproto.SCAck{ScHuAck: huAck}
	m.game.Send2Player(ack, game.SeatAll)
}

func (m *Messager) sendDrawAck(tile int32) {
	drawAck := &scproto.SCDrawAck{
		Seat: m.play.GetCurSeat(),
		Tile: tile,
	}
	ack := &scproto.SCAck{ScDrawAck: drawAck}
	m.game.Send2Player(ack, drawAck.Seat)
	drawAck.Seat = mahjong.SeatNull
	for i := range m.game.GetPlayerCount() {
		if i != drawAck.Seat {
			m.game.Send2Player(ack, i)
		}
	}
}

func (m *Messager) sendResult(isLiuJu bool, paoSeat, paoCiSeat int32) {
	// 实现发送麻将结果
}
