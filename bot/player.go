package bot

import (
	"math/rand"

	"github.com/kevin-chtw/tw_common/gamebase/game"
	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/kevin-chtw/tw_common/utils"
	"github.com/kevin-chtw/tw_mjsc_svr/ai"
	"github.com/kevin-chtw/tw_proto/cproto"
	"github.com/kevin-chtw/tw_proto/game/pbmj"
	"github.com/kevin-chtw/tw_proto/game/pbsc"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Player struct {
	*game.BotPlayer
	handlers    map[string]func(proto.Message) error
	gameState   *ai.GameState
	pendingReqs []struct {
		req   *pbmj.MJRequestReq
		delay int // 剩余延迟(ms)
	}
}

func NewPlayer(uid string, matchid, tableid int32) *game.BotPlayer {
	p := &Player{
		BotPlayer: game.NewBotPlayer(uid, matchid, tableid),
		handlers:  make(map[string]func(proto.Message) error),
		gameState: ai.NewGameState(),
	}

	p.Bot = p
	p.init()
	return p.BotPlayer
}

func (p *Player) init() {
	p.handlers[utils.TypeUrl(&pbmj.MJGameStartAck{})] = p.gameStartAck
	p.handlers[utils.TypeUrl(&pbmj.MJOpenDoorAck{})] = p.openDoorAck
	p.handlers[utils.TypeUrl(&pbsc.SCSwapTilesAck{})] = p.swapTileAck
	p.handlers[utils.TypeUrl(&pbsc.SCSwapFinishAck{})] = p.swapFinishAck
	p.handlers[utils.TypeUrl(&pbsc.SCSwapTilesResultAck{})] = p.swapResultAck
	p.handlers[utils.TypeUrl(&pbsc.SCDingQueAck{})] = p.dingQueAck
	p.handlers[utils.TypeUrl(&pbsc.SCDingQueResultAck{})] = p.dingQueResultAck
	p.handlers[utils.TypeUrl(&pbmj.MJRequestAck{})] = p.requestAck
	p.handlers[utils.TypeUrl(&pbmj.MJDiscardAck{})] = p.discardAck
	p.handlers[utils.TypeUrl(&pbmj.MJDrawAck{})] = p.drawAck
	p.handlers[utils.TypeUrl(&pbmj.MJHuAck{})] = p.huAck
	p.handlers[utils.TypeUrl(&pbmj.MJKonAck{})] = p.konAck
	p.handlers[utils.TypeUrl(&pbmj.MJPonAck{})] = p.ponAck
	p.handlers[utils.TypeUrl(&pbmj.MJResultAck{})] = p.resultAck
}

func (p *Player) OnTimer() error {
	// 处理待发送请求
	for i := 0; i < len(p.pendingReqs); i++ {
		p.pendingReqs[i].delay -= 1000 // 每次减少1秒
		if p.pendingReqs[i].delay <= 0 {
			if err := p.sendMsg(p.pendingReqs[i].req); err != nil {
				logger.Log.Errorf("发送请求失败: %v", err)
			}
			// 移除已发送请求
			p.pendingReqs = append(p.pendingReqs[:i], p.pendingReqs[i+1:]...)
			i-- // 调整索引
		}
	}
	return nil
}

func (p *Player) OnBotMsg(msg proto.Message) error {
	ack := msg.(*cproto.GameAck)
	gameAck, err := ack.Ack.UnmarshalNew()
	if err != nil {
		return err
	}
	if ack.Ack.TypeUrl == utils.TypeUrl(&cproto.TablePlayerAck{}) {
		playerAck := gameAck.(*cproto.TablePlayerAck)
		if playerAck.Uid == p.Uid {
			p.gameState.CurrentSeat = int(playerAck.Seat)
		}
		return nil
	}
	if ack.Ack.TypeUrl == utils.TypeUrl(&cproto.TableMsgAck{}) {
		payLoads := gameAck.(*cproto.TableMsgAck).GetMsg()
		scAck := &pbsc.SCAck{}
		if err := proto.Unmarshal(payLoads, scAck); err != nil {
			return err
		}
		inMsg, err := scAck.Ack.UnmarshalNew()
		if err != nil {
			return err
		}
		h, ok := p.handlers[scAck.Ack.TypeUrl]
		if ok {
			return h(inMsg)
		}
	}
	return nil
}

func (p *Player) sendMsg(msg proto.Message) error {
	data, err := anypb.New(msg)
	if err != nil {
		return err
	}
	req := &pbsc.SCReq{Req: data}
	return p.SendMsg(req)
}

func (p *Player) gameStartAck(msg proto.Message) error {
	return nil
}

func (p *Player) openDoorAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJOpenDoorAck)
	if ack.Seat != int32(p.gameState.CurrentSeat) {
		for _, tile := range ack.GetTiles() {
			p.gameState.Hand[mahjong.Tile(tile)]++
		}
		p.gameState.TotalTiles -= (13*4 + 1)
		return nil
	}
	return nil
}

func (p *Player) swapTileAck(msg proto.Message) error {
	return nil
}

func (p *Player) swapFinishAck(msg proto.Message) error {
	ack := msg.(*pbsc.SCSwapFinishAck)
	if ack.Seat == int32(p.gameState.CurrentSeat) {
		for _, tile := range ack.GetTiles() {
			p.gameState.Hand[mahjong.Tile(tile)]--
		}
	}
	return nil
}

func (p *Player) swapResultAck(msg proto.Message) error {
	ack := msg.(*pbsc.SCSwapTilesResultAck)
	swaps := ack.GetSwapTiles()
	for _, t := range swaps[p.gameState.CurrentSeat].Tiles {
		p.gameState.Hand[mahjong.Tile(t)]++
	}

	return nil
}

func (p *Player) dingQueAck(msg proto.Message) error {
	return nil
}

func (p *Player) dingQueResultAck(msg proto.Message) error {
	ack := msg.(*pbsc.SCDingQueResultAck)
	for i, c := range ack.GetColors() {
		p.gameState.PlayerLacks[i] = mahjong.EColor(c)
	}
	return nil

}

func (p *Player) requestAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJRequestAck)
	if ack.Seat != int32(p.gameState.CurrentSeat) {
		return nil
	}
	p.gameState.Operates = mahjong.NewOperates(ack.RequestType)
	ret := ai.GetRichAI(true).Step(p.gameState)
	req := &pbmj.MJRequestReq{
		Seat:        ack.Seat,
		RequestType: int32(ret.Operate),
		Requestid:   ack.Requestid,
		Tile:        ret.Tile.ToInt32(),
	}

	// 添加到待发送队列
	delay := 1000 + rand.Intn(1000) // 3-5秒随机延迟
	p.pendingReqs = append(p.pendingReqs, struct {
		req   *pbmj.MJRequestReq
		delay int
	}{
		req:   req,
		delay: delay,
	})
	return nil
}

// 处理各种ACK并记录操作
func (p *Player) discardAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJDiscardAck)

	if p.gameState.PlayerMelds[int(ack.Seat)] == nil {
		p.gameState.PlayerMelds[int(ack.Seat)] = make(map[mahjong.Tile]int)
	}

	p.gameState.PlayerMelds[int(ack.Seat)][mahjong.Tile(ack.Tile)]++
	p.gameState.LastTile = mahjong.Tile(ack.Tile)
	if ack.Seat == int32(p.gameState.CurrentSeat) {
		p.gameState.Hand[mahjong.Tile(ack.Tile)]--
	}

	return nil
}

func (p *Player) drawAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJDrawAck)
	p.gameState.TotalTiles--
	if ack.Seat == int32(p.gameState.CurrentSeat) {
		p.gameState.Hand[mahjong.Tile(ack.Tile)]++
	}
	return nil
}

func (p *Player) huAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJHuAck)
	for _, h := range ack.HuData {
		p.gameState.HuPlayers = append(p.gameState.HuPlayers, int(h.Seat))
	}
	return nil
}

func (p *Player) konAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJKonAck)
	p.gameState.GangTiles[int(ack.Seat)] = append(p.gameState.GangTiles[int(ack.Seat)], mahjong.Tile(ack.Tile))
	return nil
}
func (p *Player) ponAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJPonAck)
	p.gameState.PonTiles[int(ack.Seat)] = append(p.gameState.PonTiles[int(ack.Seat)], mahjong.Tile(ack.Tile))
	return nil
}

func (p *Player) resultAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJResultAck)

	// 计算最终得分
	var finalScore float32
	for _, player := range ack.PlayerResults {
		if player.Seat == int32(p.gameState.CurrentSeat) {
			finalScore = float32(player.WinScore)
			break
		}
	}
	isWin := finalScore > 0
	ai.GetRichAI(true).NotifyGameResult(p.gameState, isWin, finalScore)
	return nil
}
