package mjsc

const (
	RuleDiscardTime = iota //出牌时间 0
	RuleWaitTime           //等待时间 1
	RuleSwapTile           //换三张 2
	RuleZiMoJiaDi          //自摸加底 3
	RuleMaxMulti           //封顶倍数 4
	RuleTianDiHu           //天地胡 5
	RuleJiangDui19         //幺九将对 6
	RuleMQZZ               //门清中张 7
	RuleYiTiaoLong         //一条龙 8
	RuleJiaXinWu           //夹心五 9
	RuleKaBianZhang        //卡边张 10
	RuleZhuanYu            //呼叫转移 11
	RuleChaJiao            //查大叫 12
	RuleTuiYu              //退雨 13
	RuleHaiDi              //海底捞月 14
	RuleEnd
)
