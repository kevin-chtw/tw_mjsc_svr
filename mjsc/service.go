package mjsc

import (
	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/kevin-chtw/tw_proto/game/pbmj"
)

func init() {
	mahjong.Service = NewService()
}

type service struct {
	tiles        map[mahjong.Tile]int
	tiles2Men    map[mahjong.Tile]int
	defaultRules [RuleEnd]int
	huCore       *mahjong.HuCore
	fdRules      map[string]int32
}

func NewService() mahjong.IService {
	s := &service{
		tiles:        make(map[mahjong.Tile]int),
		tiles2Men:    make(map[mahjong.Tile]int),
		defaultRules: [RuleEnd]int{10, 8, 0, 1, 10, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 1, 0, 1},
		huCore:       mahjong.NewHuCore(14),
		fdRules:      make(map[string]int32),
	}
	s.initTiles()
	s.initFdRules()
	return s
}
func (s *service) initTiles() {
	for color := mahjong.ColorCharacter; color <= mahjong.ColorDot; color++ {
		pc := mahjong.PointCountByColor[color]
		for i := range pc {
			tile := mahjong.MakeTile(color, i)
			s.tiles[tile] = 4
			if color != mahjong.ColorCharacter {
				s.tiles2Men[tile] = 4
			}
		}
	}
}

func (s *service) GetAllTiles(conf *mahjong.Rule) map[mahjong.Tile]int {
	if conf.GetValue(RuleLiangMen) != 0 {
		return s.tiles2Men
	} else {
		return s.tiles
	}
}

func (s *service) GetHandCount() int {
	return 13
}

func (s *service) GetDefaultRules() []int {
	return s.defaultRules[:]
}
func (s *service) initFdRules() {
	s.fdRules["huansz"] = RuleSwapTile       //换三张
	s.fdRules["zimojd"] = RuleZiMoJiaDi      //自摸加底
	s.fdRules["maxmulti"] = RuleMaxMulti     //封顶倍数
	s.fdRules["tiandihu"] = RuleTianDiHu     //天地胡
	s.fdRules["jiangdui19"] = RuleJiangDui19 //幺九将对
	s.fdRules["mqzz"] = RuleMQZZ             //门清中张
	s.fdRules["yitiaolong"] = RuleYiTiaoLong //一条龙
	s.fdRules["jiaxw"] = RuleJiaXinWu        //夹心五
	s.fdRules["kabianz"] = RuleKaBianZhang   //卡边张
	s.fdRules["zhuanyu"] = RuleZhuanYu       //呼叫转移
	s.fdRules["chajiao"] = RuleChaJiao       //查大叫
	s.fdRules["tuiyu"] = RuleTuiYu           //退雨
	s.fdRules["haidi"] = RuleHaiDi           //海底捞月
	s.fdRules["liangmen"] = RuleLiangMen     //两门
	s.fdRules["sijbsj"] = RuleSiJBSJ         //死叫不算叫
	s.fdRules["chagua"] = RuleChaGua         //擦挂
	s.fdRules["diankhsdp"] = RuleDianKHSDP   //点杠花算点炮
	s.fdRules["juezhang"] = RuleJueZhang     //绝张
}

func (s *service) GetFdRules() map[string]int32 {
	return s.fdRules
}

func (s *service) GetHuResult(data *mahjong.HuData) *pbmj.MJHuData {
	h := newHuData(data)
	result := data.InitHuResult()
	result.Gen = h.calcGen()
	result.HuTypes = h.getHuTypes()

	result.Multi = totalMuti(result, data.Play.GetRule())
	return result
}
