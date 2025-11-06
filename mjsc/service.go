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
	tilesFeng    map[mahjong.Tile]int
	defaultRules [RuleEnd]int
	huCore       *mahjong.HuCore
	fdRules      map[string]int32
}

func NewService() mahjong.IService {
	s := &service{
		tiles:        make(map[mahjong.Tile]int),
		tilesFeng:    make(map[mahjong.Tile]int),
		defaultRules: [RuleEnd]int{10, 8, 1, 10, 36, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		huCore:       mahjong.NewHuCore(14),
		fdRules:      make(map[string]int32),
	}
	s.init()
	s.initFdRules()
	return s
}
func (s *service) init() {
	for color := mahjong.ColorCharacter; color <= mahjong.ColorDragon; color++ {
		pc := mahjong.PointCountByColor[color]
		for i := range pc {
			tile := mahjong.MakeTile(color, i)
			if color < mahjong.ColorWind {
				s.tiles[tile] = 4
			}
			s.tilesFeng[tile] = 4
		}
	}

}

func (s *service) GetAllTiles(conf *mahjong.Rule) map[mahjong.Tile]int {
	return s.tiles
}

func (s *service) GetHandCount() int {
	return 13
}

func (s *service) GetDefaultRules() []int {
	return s.defaultRules[:]
}
func (s *service) initFdRules() {
	s.fdRules["huansz"] = RuleSwapTile       //换三张 2
	s.fdRules["zimojd"] = RuleZiMoJiaDi      //自摸加底 3
	s.fdRules["maxmulti"] = RuleMaxMulti     //封顶倍数 4
	s.fdRules["tiandihu"] = RuleTianDiHu     //天地胡 5
	s.fdRules["jiangdui19"] = RuleJiangDui19 //幺九将对 6
	s.fdRules["mqzz"] = RuleMQZZ             //门清中张 7
	s.fdRules["yitiaolong"] = RuleYiTiaoLong //一条龙 8
	s.fdRules["jiaxw"] = RuleJiaXinWu        //夹心五 9
	s.fdRules["kabianz"] = RuleKaBianZhang   //卡边张 10
	s.fdRules["zhuanyu"] = RuleZhuanYu       //呼叫转移 11
	s.fdRules["chajiao"] = RuleChaJiao       //查大叫 12
	s.fdRules["tuiyu"] = RuleTuiYu           //退雨 13
	s.fdRules["haidi"] = RuleHaiDi           //海底捞月 14
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
