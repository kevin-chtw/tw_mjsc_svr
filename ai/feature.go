package ai

const (
	// HistorySteps 历史操作序列长度（最大60步，包含所有玩家的操作）
	HistorySteps = 100
	// HistoryStepDim 每步操作编码维度：操作类型(5) + 玩家座位(4) + 牌索引(34) = 43
	HistoryStepDim = 5 + 4 + 34
	// HistoryDim 历史操作序列总维度：100 * 43 = 4300
	HistoryDim = HistorySteps * HistoryStepDim
)

// RichFeature 精简特征（专注于核心信息）
type RichFeature struct {
	TotalTiles    float32             // 1 - 总牌张数（归一化）
	CurrentSeat   [4]float32          // 4 - 当前玩家座位号（one-hot编码）
	Operates      [5]float32          // 5 - 当前可执行操作（one-hot编码）
	Hand          [34]float32         // 34 - 手牌
	PlayerLacks   [4][3]float32       // 4×3 - 各玩家缺门花色（one-hot编码）
	ActionHistory [HistoryDim]float32 // 历史操作序列（最多100步，每步43维：操作类型5+玩家座位4+牌索引34）
}

// ToVector flatten → []float32 精简特征向量
func (f RichFeature) ToVector() []float32 {
	// 计算新特征维度: 56 + 4300 = 4356
	// 基础特征(605) + 历史操作序列(100×43=4300)
	out := make([]float32, 0, 56+HistoryDim)
	out = append(out, f.TotalTiles)        // 1
	out = append(out, f.CurrentSeat[:]...) // 4
	out = append(out, f.Operates[:]...)    // 5
	out = append(out, f.Hand[:]...)        // 34
	for i := range 4 {
		out = append(out, f.PlayerLacks[i][:]...) // 4×3 = 12
	}
	// 历史操作序列 (4300)
	out = append(out, f.ActionHistory[:]...) // 100×43 = 4300

	return out
}
