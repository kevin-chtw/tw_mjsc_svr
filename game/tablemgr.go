package game

import (
	"sync"

	pitaya "github.com/topfreegames/pitaya/v3/pkg"
)

// TableManager 管理游戏桌
type TableManager struct {
	mu     sync.RWMutex
	tables map[string]*Table // tableID -> Table
	app    pitaya.Pitaya
}

// NewTableManager 创建游戏桌管理器
func NewTableManager(app pitaya.Pitaya) *TableManager {
	return &TableManager{
		tables: make(map[string]*Table),
		app:    app,
	}
}

// GetTable 获取指定比赛和桌号的游戏桌
func (tm *TableManager) Get(matchID, tableID string) *Table {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	key := matchID + ":" + tableID
	return tm.tables[key]
}

// LoadOrStore 加载或存储游戏桌
func (tm *TableManager) LoadOrStore(matchID, tableID string) *Table {
	key := matchID + ":" + tableID

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 检查是否已存在
	if table, ok := tm.tables[key]; ok {
		return table
	}

	// 创建新表
	table := NewTable(matchID, tableID, tm.app)
	tm.tables[key] = table
	return table
}

// Delete 删除指定比赛和桌号的游戏桌
func (tm *TableManager) Delete(matchID, tableID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	key := matchID + ":" + tableID
	delete(tm.tables, key)
}
