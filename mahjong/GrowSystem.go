package mahjong

// GrowSystem 成长系统
type GrowSystem struct {
	Level int
	Exp   int
}

// NewGrowSystem 创建新的成长系统
func NewGrowSystem() *GrowSystem {
	return &GrowSystem{
		Level: 1,
		Exp:   0,
	}
}

// AddExp 增加经验值
func (g *GrowSystem) AddExp(exp int) {
	g.Exp += exp
	// 实现升级逻辑
}

// GetLevel 获取当前等级
func (g *GrowSystem) GetLevel() int {
	return g.Level
}

// GetExp 获取当前经验值
func (g *GrowSystem) GetExp() int {
	return g.Exp
}