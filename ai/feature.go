package ai

// RichFeature 精简特征（专注于核心信息）
type RichFeature struct {
	Hand          [34]float32    // 34 - 手牌
	Furo          [4][34]float32 // 4×34 - 各玩家副露
	PonInfo       [4][34]float32 // 4×34 - 各玩家碰牌信息
	GangInfo      [4][34]float32 // 4×34 - 各玩家杠牌信息
	HuInfo        [4][34]float32 // 4×34 - 各玩家胡牌信息
	PlayerActions [4]float32     // 4 - 玩家动作状态
	PlayerLacks   [4][3]float32  // 4×3 - 各玩家缺门花色（one-hot编码）
	CurrentSeat   [4]float32     // 4 - 当前玩家座位号（one-hot编码）
	TotalTiles    float32        // 1 - 总牌张数（归一化）
	Operates      [5]float32     // 5 - 当前可执行操作（one-hot编码）
}

// ToVector flatten → []float32 精简特征向量
func (f RichFeature) ToVector() []float32 {
	// 计算新特征维度: 34 + 136 + 408 + 21 + 5 = 604
	// 手牌(34) + 副露(4×34=136) + 碰杠胡(3×4×34=408) + 玩家状态(4+12+4+1=21) + 操作(5)
	out := make([]float32, 0, 604)

	// 手牌和副露特征 (34 + 136 = 170)
	out = append(out, f.Hand[:]...) // 34
	for i := range 4 {
		out = append(out, f.Furo[i][:]...) // 4×34 = 136
	}

	// 碰杠胡特征 (136 + 136 + 136 = 408)
	for i := range 4 {
		out = append(out, f.PonInfo[i][:]...)  // 4×34 = 136
		out = append(out, f.GangInfo[i][:]...) // 4×34 = 136
		out = append(out, f.HuInfo[i][:]...)   // 4×34 = 136
	}

	// 玩家状态特征 (4 + 12 + 4 + 1 = 21)
	out = append(out, f.PlayerActions[:]...) // 4
	for i := range 4 {
		out = append(out, f.PlayerLacks[i][:]...) // 4×3 = 12
	}
	out = append(out, f.CurrentSeat[:]...) // 4
	out = append(out, f.TotalTiles)        // 1

	// 操作特征 (5)
	out = append(out, f.Operates[:]...) // 5

	return out
}

// DangerMask 危险度 0=安全 1=危险
func DangerMask(f *RichFeature, lack int) [34]bool {
	var mask [34]bool
	for i := range 34 {
		suit := i / 9
		if suit == lack {
			mask[i] = true // 缺门→别人可能要
			continue
		}
		seen := 0
		for p := range 4 {
			seen += int(f.Furo[p][i])
		}
		mask[i] = seen < 2 // 已见<2 张→危险
	}
	return mask
}

func EffiReward(old, new [34]float32, lack int) float32 {
	return 0
}
