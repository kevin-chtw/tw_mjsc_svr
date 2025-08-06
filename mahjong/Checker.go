package mahjong

import (
)

// ScoreNode 分数节点
type ScoreNode struct {
	WinScoreType ScoreType
	Multiple     []int64
	ScoresOrigin []int64
	ScoresFinal  []int64
}

// Checker 麻将分数检查器基类
type Checker struct {
	game      *Game
	scoreType EScoreType
}

// NewChecker 创建新的分数检查器
func NewChecker(game *Game, minScoreType EScoreType) *Checker {
	return &Checker{
		game:      game,
		scoreType: minScoreType,
	}
}

// CheckBankrupt 检查破产玩家
func (c *Checker) CheckBankrupt() []ISeatID {
	// 实现检查逻辑
	return nil
}

// SetRestLoseOut 设置剩余玩家为输家
func (c *Checker) SetRestLoseOut() []ISeatID {
	// 实现设置逻辑
	return nil
}

// SetBankruptOut 设置破产玩家
func (c *Checker) SetBankruptOut() []ISeatID {
	// 实现设置逻辑
	return nil
}

// SyncAllPlayerScore 同步所有玩家分数
func (c *Checker) SyncAllPlayerScore() {
	// 实现同步逻辑
}

// CalFinalScore 计算最终分数
func (c *Checker) CalFinalScore(takeScores, oriScores []int64, limited *[]bool) []int64 {
	if limited == nil {
		return c.calcMinScore(takeScores, oriScores, nil)
	}
	return c.calcMinScore(takeScores, oriScores, *limited)
}

func (c *Checker) calcMinScore(takeScores, oriScores []int64, limited []bool) []int64 {
	// 实现最小分数计算逻辑
	return nil
}

// CheckerMany 多次计分检查器
type CheckerMany struct {
	*Checker
	bills []ScoreNode
}

// NewCheckerMany 创建多次计分检查器
func NewCheckerMany(game *Game, minScoreType EScoreType) *CheckerMany {
	return &CheckerMany{
		Checker: NewChecker(game, minScoreType),
	}
}

// Check 检查分数
func (cm *CheckerMany) Check(scoreType ScoreType, multiples []int64, pOrigins *[]int64, pLimited *[]bool) []int64 {
	// 实现检查逻辑
	return nil
}

// CheckWithScores 根据分数检查
func (cm *CheckerMany) CheckWithScores(scoreType ScoreType, origins []int64, pLimited *[]bool) []int64 {
	// 实现检查逻辑
	return nil
}

// GetFullBills 获取完整账单
func (cm *CheckerMany) GetFullBills() []ScoreNode {
	return cm.bills
}

// RecordScoreInfo 记录分数信息
func (cm *CheckerMany) RecordScoreInfo(scoreType ScoreType, multiples, origins, finals []int64, pLimited *[]bool) {
	// 实现记录逻辑
}

// CheckerOnce 单次计分检查器
type CheckerOnce struct {
	*Checker
	multiplesMap   map[ScoreType][]int64
	totalMultiples []int64
}

// NewCheckerOnce 创建单次计分检查器
func NewCheckerOnce(game *Game, minScoreType EScoreType) *CheckerOnce {
	return &CheckerOnce{
		Checker:    NewChecker(game, minScoreType),
		multiplesMap: make(map[ScoreType][]int64),
	}
}

// AddMultiple 添加倍数
func (co *CheckerOnce) AddMultiple(scoreType ScoreType, multiple []int64) {
	co.multiplesMap[scoreType] = multiple
}

// Check 检查分数
func (co *CheckerOnce) Check(pOrigins *[]int64, pLimited *[]bool) []int64 {
	// 实现检查逻辑
	return nil
}

// GetMultiples 获取倍数
func (co *CheckerOnce) GetMultiples(scoreType ScoreType) []int64 {
	return co.multiplesMap[scoreType]
}

// GetMultiple 获取玩家倍数
func (co *CheckerOnce) GetMultiple(seat ISeatID, scoreType ScoreType) int64 {
	return co.multiplesMap[scoreType][seat]
}

// GetOriginScore 获取原始分数
func (co *CheckerOnce) GetOriginScore(seat ISeatID, scoreType ScoreType) int64 {
	return co.multiplesMap[scoreType][seat]
}

// GetTotalMultiples 获取总倍数
func (co *CheckerOnce) GetTotalMultiples() []int64 {
	return co.totalMultiples
}

// ScoreValueChange 分数变化计算
func ScoreValueChange(result []int64, base int64, winSeat ISeatID, relatedSeats []ISeatID) {
	for _, seat := range relatedSeats {
		if seat != winSeat {
			result[seat] -= base
			result[winSeat] += base
		}
	}
}

// ScoreValueChangeWithMultiple 带倍数的分数变化计算
func ScoreValueChangeWithMultiple(result []int64, base int64, winSeat ISeatID, relatedSeats []ISeatID, multiple []int64) {
	m1 := multiple[winSeat]
	for _, seat := range relatedSeats {
		if seat != winSeat {
			score := base * m1 * multiple[seat]
			result[seat] -= score
			result[winSeat] += score
		}
	}
}