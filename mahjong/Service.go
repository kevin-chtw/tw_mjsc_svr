package mahjong

type GameCounter struct {
	gameCount int32
}

var gameCounterInstance *GameCounter

func GetGameCounter() *GameCounter {
	if gameCounterInstance == nil {
		gameCounterInstance = &GameCounter{}
	}
	return gameCounterInstance
}

func (c *GameCounter) AddGameCount() {
	c.gameCount++
}

func (c *GameCounter) DecGameCount() {
	c.gameCount--
}

func (c *GameCounter) GetGameCount() int {
	return int(c.gameCount)
}

type HuBase struct{}

func (h *HuBase) CheckCall(playerInfo *HuPlayData, cfg *Config) map[int]map[int]int {
	return make(map[int]map[int]int)
}

func (h *HuBase) CheckHu(result *HuResult, playerInfo *HuPlayData, cfg *Config) bool {
	return false
}

type Service struct {
	version    string
	moduleName string
	tingTool   *TingBase
	huTool     *HuBase
}

func NewService() *Service {
	return &Service{
		version:    loadModuleVersion(),
		moduleName: loadModuleName(),
	}
}

func (s *Service) OnInitialUpdate() bool {
	s.tingTool = s.OnCreateTingTool()
	s.huTool = s.OnCreateHuTool()
	return true
}

func (s *Service) OnCreateTingTool() *TingBase {
	return NewTingEx14()
}

func (s *Service) OnCreateHuTool() *HuBase {
	return &HuBase{}
}

func (s *Service) GetTingTool() *TingBase {
	return s.tingTool
}

func (s *Service) GetHuTool() *HuBase {
	return s.huTool
}

func (s *Service) GetHandCount() int {
	return TileCountInitNormal
}

func (s *Service) GetVersion() string {
	return s.version
}

func (s *Service) GetModuleName() string {
	return s.moduleName
}

type MJService struct {
	*Service
	*GameCounter
	gameID int
	port   int
}

func NewMJService[TGame any, TPlayer any](gameID, port int) *MJService {
	return &MJService{
		Service:     NewService(),
		GameCounter: GetGameCounter(),
		gameID:      gameID,
		port:        port,
	}
}

func (s *MJService) InitService() {
	// 实现服务初始化逻辑
}

func (s *MJService) OnRegisterTKObject() bool {
	// 实现注册TK对象逻辑
	return true
}

func (s *MJService) OnCreateGame() interface{} {
	// 实现创建游戏逻辑
	return nil
}

func (s *MJService) OnCreateTKGamePlayer() interface{} {
	// 实现创建游戏玩家逻辑
	return nil
}

func GetMJService() *Service {
	// 实现获取MJService实例逻辑
	return nil
}

func loadModuleVersion() string {
	// 实现加载模块版本逻辑
	return ""
}

func loadModuleName() string {
	// 实现加载模块名称逻辑
	return ""
}
