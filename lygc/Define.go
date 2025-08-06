package lygc

import "github.com/kevin-chtw/tw_game_svr/mahjong"

const (
	UseOldDraw     = true
	MaxPlayerCount = 4
)

const (
	OperateCI = 1 << 11 // 次牌操作
)

type MJTips int

const (
	TipPassHuLong MJTips = 12 // 过胡龙
	TipTing       MJTips = 13 // 听牌状态
	TipBaoCiState MJTips = 14 // 包次状态
)

type MJScoreType int

const (
	ScoreTypeGang      MJScoreType = 3
	ScoreTypeGenZhuang MJScoreType = 4
	ScoreTypeMi        MJScoreType = 5
)

type ECallType int

const (
	CallTypeNone      ECallType = 0
	CallTypeCall      ECallType = 1
	CallTypeFirstCall ECallType = 2
)

type RuleItem int

const (
	RuleItemMannualCardsIndex RuleItem = iota // 手动牌索引
	RuleItemDiscardTime                       // 弃牌时间
	RuleItemWaitTime                          // 等待时间

	RuleItemScoreType         // 计分方式
	RuleItemExpression        // 表情
	RuleItemShortCode         // 短码
	RuleItemRuleFeng          // 风牌规则
	RuleItemCiPai             // 次牌
	RuleItemGenZhuang         // 跟庄
	RuleItemCiHuBei           // 次胡倍
	RuleItemPiCiBei           // 皮次倍
	RuleItemXiaMi             // 下密
	RuleItemXiaMiBei          // 下密倍
	RuleItemDrawCtrl          // 抽牌控制
	RuleItemDealerCtrl        // 庄家控制
	RuleItemDispatchXiangTing // 调度项听
	RuleItemMoBomb            // 摸炸弹
	RuleItemDealerAnKe        // 庄家暗刻
	RuleItemDealerAnGang      // 庄家暗杠
	RuleItemDealerDoubleAnKe  // 庄家双暗刻
	RuleItemDealerDuiZi       // 庄家对子
	RuleItemEnd
)

var DefaultConfigValues = []int{
	0, 10, 8, 1, 1, 1, 0, 0, 0, 3, 3, 0, 3, 0, 0, 3, 0, 22, 10, 8, 10,
}

var FDTableConfigMap = map[string]RuleItem{
	"DuanYu":       RuleItemShortCode,
	"BiaoQing":     RuleItemExpression,
	"RuleFeng":     RuleItemRuleFeng,
	"CiPai":        RuleItemCiPai,
	"CiHuBei":      RuleItemCiHuBei,
	"PiCiBei":      RuleItemPiCiBei,
	"XiaMi":        RuleItemXiaMi,
	"XiaMiBei":     RuleItemXiaMiBei,
	"GenZhuang":    RuleItemGenZhuang,
	"YingCiZhaDan": RuleItemMoBomb,
}

type LastHandData struct {
	Banker        mahjong.ISeatID
	FirstBanker   mahjong.ISeatID
	LianZhuang    int // 连庄次数
	LastXiaMiSeat mahjong.ISeatID
	LastHuType    int
	MiChi         int
	MiCount       [4]int
}

func NewLastHandData() *LastHandData {
	return &LastHandData{
		Banker:        mahjong.SeatNull,
		FirstBanker:   mahjong.SeatNull,
		LastXiaMiSeat: mahjong.SeatNull,
		LastHuType:    -1,
	}
}
