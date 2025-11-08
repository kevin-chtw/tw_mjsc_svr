package mjsc

const (
	RuleDiscardTime = iota //出牌时间 0
	RuleWaitTime           //等待时间 1
	RuleScoreType          //算分方式 2
	RuleSwapTile           //换三张 3
	RuleZiMoJiaDi          //自摸加底 4
	RuleMaxMulti           //封顶倍数 5
	RuleTianDiHu           //天地胡 6
	RuleJiangDui19         //幺九将对 7
	RuleMQZZ               //门清中张 8
	RuleYiTiaoLong         //一条龙 9
	RuleJiaXinWu           //夹心五 10
	RuleKaBianZhang        //卡边张 11
	RuleZhuanYu            //呼叫转移 12
	RuleChaJiao            //查大叫 13
	RuleTuiYu              //退雨 14
	RuleHaiDi              //海底捞月 15
	RuleLiangMen           //两门 16
	RuleEnd
)
