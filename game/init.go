package game

import pitaya "github.com/topfreegames/pitaya/v3/pkg"

var playerManager *PlayerManager
var tableManager *TableManager

// InitGame 初始化游戏模块
func InitGame(app pitaya.Pitaya) {
	playerManager = NewPlayerManager()
	tableManager = NewTableManager(app)
}

// GetPlayerManager 获取玩家管理器实例
func GetPlayerManager() *PlayerManager {
	return playerManager
}

// GetTableManager 获取游戏桌管理器实例
func GetTableManager() *TableManager {
	return tableManager
}
