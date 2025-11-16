package bot

import (
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
	pendingReqs []*game.PendingReq
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
	p.handlers[utils.TypeUrl(&pbmj.MJAnimationAck{})] = p.animationAck
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
		p.pendingReqs[i].Delay -= 1000 // 每次减少1秒
		if p.pendingReqs[i].Delay <= 0 {
			if err := p.sendMsg(p.pendingReqs[i].Req); err != nil {
				logger.Log.Errorf("发送请求失败: %v", err)
			}
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
			p.Seat = playerAck.Seat
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

func (p *Player) delayMsg(req proto.Message) {
	// 训练模式下立即发送，不延迟
	if ai.IsTrainingMode() {
		p.sendMsg(req)
		return
	}

	// 生产模式：添加到待发送队列
	delay := 0 //1000 + rand.Intn(1000) // 3-5秒随机延迟
	p.pendingReqs = append(p.pendingReqs, &game.PendingReq{
		Req:   req,
		Delay: delay,
	})
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
	p.gameState = ai.NewGameState()
	p.gameState.CurrentSeat = int(p.Seat)
	return nil
}

func (p *Player) openDoorAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJOpenDoorAck)
	if ack.Seat == int32(p.gameState.CurrentSeat) {
		for _, tile := range ack.GetTiles() {
			p.gameState.Hand[mahjong.Tile(tile)]++
		}
		p.gameState.TotalTiles -= (13*4 + 1)
	}
	logger.Log.Infof("seat=%d, %v", p.gameState.CurrentSeat, p.gameState.Hand)
	return nil
}

func (p *Player) animationAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJAnimationAck)
	req := &pbmj.MJAnimationReq{
		Seat:      int32(p.gameState.CurrentSeat),
		Requestid: ack.Requestid,
	}
	p.delayMsg(req)
	return nil
}

func (p *Player) swapTileAck(msg proto.Message) error {
	ack := msg.(*pbsc.SCSwapTilesAck)
	colorCount := make(map[mahjong.EColor]int)            // key: 花色, value: 牌数
	colorTiles := make(map[mahjong.EColor][]mahjong.Tile) // key: 花色, value: 该花色的牌

	for tile, count := range p.gameState.Hand {
		color := tile.Color()
		colorCount[color] += count
		colorTiles[color] = append(colorTiles[color], mahjong.MakeTiles(tile, count)...)
	}

	// 找出牌数大于3张且张数最少的花色
	var bestColor mahjong.EColor = mahjong.ColorUndefined
	var minCount int = 999
	for color, count := range colorCount {
		if count > 3 && count < minCount {
			bestColor = color
			minCount = count
		}
	}

	if bestColor == mahjong.ColorUndefined {
		return nil
	}
	tiles := colorTiles[bestColor]
	req := &pbsc.SCSwapTilesReq{
		Requestid: ack.Requestid,
		Tiles:     mahjong.TilesInt32(tiles[:3]),
	}
	p.delayMsg(req)
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
	for _, s := range swaps {
		if s.To == int32(p.gameState.CurrentSeat) {
			for _, t := range s.Tiles {
				p.gameState.Hand[mahjong.Tile(t)]++
			}
		}
	}

	logger.Log.Infof("seat=%d, %v", p.gameState.CurrentSeat, p.gameState.Hand)
	return nil
}

func (p *Player) dingQueAck(msg proto.Message) error {
	ack := msg.(*pbsc.SCDingQueAck)
	colors := make(map[mahjong.EColor]int32)
	for tile, count := range p.gameState.Hand {
		colors[tile.Color()] += int32(count)
	}
	bestColor := mahjong.ColorCharacter
	min := colors[mahjong.ColorCharacter]
	for c := mahjong.ColorCharacter + 1; c <= mahjong.ColorDot; c++ {
		if count, ok := colors[c]; !ok {
			min = 0
			bestColor = c
		} else if count < min {
			min = count
			bestColor = c
		}
	}
	req := &pbsc.SCDingQueReq{
		Requestid: ack.Requestid,
		Color:     int32(bestColor),
	}
	p.delayMsg(req)
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
	logger.Log.Info(p.gameState.Hand)
	p.gameState.Operates = mahjong.NewOperates(ack.RequestType)
	ret := ai.GetRichAI().Step(p.gameState)
	req := &pbmj.MJRequestReq{
		Seat:        ack.Seat,
		RequestType: int32(ret.Operate),
		Requestid:   ack.Requestid,
		Tile:        ret.Tile.ToInt32(),
	}
	p.delayMsg(req)
	return nil
}

// 处理各种ACK并记录操作
func (p *Player) discardAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJDiscardAck)

	p.gameState.LastTile = mahjong.Tile(ack.Tile)
	if ack.Seat == int32(p.gameState.CurrentSeat) {
		p.gameState.Hand[mahjong.Tile(ack.Tile)]--
	}

	// 记录实现操作（出牌）
	p.gameState.RecordAction(int(ack.Seat), mahjong.OperateDiscard, mahjong.Tile(ack.Tile))

	return nil
}

func (p *Player) drawAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJDrawAck)
	p.gameState.TotalTiles--
	if ack.Seat == int32(p.gameState.CurrentSeat) {
		p.gameState.LastTile = mahjong.Tile(ack.Tile)
		p.gameState.Hand[mahjong.Tile(ack.Tile)]++
		p.gameState.CallData = ack.CallData
	}
	return nil
}

func (p *Player) huAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJHuAck)
	for _, h := range ack.HuData {
		p.gameState.HuPlayers = append(p.gameState.HuPlayers, int(h.Seat))
		p.gameState.HuMultis[int(h.Seat)] = h.Multi // 记录番数
		p.gameState.RecordAction(int(h.Seat), mahjong.OperateHu, mahjong.Tile(ack.Tile))
	}
	return nil
}

func (p *Player) konAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJKonAck)
	p.gameState.KonTiles[int(ack.Seat)] = append(p.gameState.KonTiles[int(ack.Seat)], mahjong.Tile(ack.Tile))
	if ack.Seat == int32(p.gameState.CurrentSeat) {
		p.gameState.Hand[mahjong.Tile(ack.Tile)] = 0
	}
	// 记录实现操作（杠牌）
	p.gameState.RecordAction(int(ack.Seat), mahjong.OperateKon, mahjong.Tile(ack.Tile))
	return nil
}
func (p *Player) ponAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJPonAck)
	p.gameState.PonTiles[int(ack.Seat)] = append(p.gameState.PonTiles[int(ack.Seat)], mahjong.Tile(ack.Tile))
	if ack.Seat == int32(p.gameState.CurrentSeat) {
		p.gameState.Hand[mahjong.Tile(ack.Tile)] -= 2
		p.gameState.CallData = ack.CallData
	}
	// 记录实现操作（碰牌）
	p.gameState.RecordAction(int(ack.Seat), mahjong.OperatePon, mahjong.Tile(ack.Tile))
	return nil
}

func (p *Player) resultAck(msg proto.Message) error {
	ack := msg.(*pbmj.MJResultAck)

	// 设置终局信息到 GameState
	for _, player := range ack.PlayerResults {
		if player.Seat == int32(p.gameState.CurrentSeat) {
			p.gameState.FinalScore = float32(player.WinScore)
			break
		}
	}

	ai.GetRichAI().QueueTraining(p.gameState)
	return nil
}
