package tools

import (
	"path/filepath"
	"testing"

	"easy-text/backend/config"
	"gorm.io/gorm"
)

// setupDBTest 为需要 SQLite 的用例创建独立临时库，返回清理函数。
// 结束时必须关闭连接，否则 Windows 上 t.TempDir 因文件句柄未释放而删除失败。
func setupDBTest(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	if err := config.InitDatabase(filepath.Join(tmpDir, "test.db")); err != nil {
		t.Fatalf("InitDatabase: %v", err)
	}
	return func() { closeDB(config.DB) }
}

func closeDB(db *gorm.DB) {
	if db == nil {
		return
	}
	if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
		_ = sqlDB.Close()
	}
}

// intToStr 供并发用例生成不重复的文件名后缀。
func intToStr(i int) string {
	if i == 0 {
		return "0"
	}
	return intToStr(i/10) + string(rune('0'+i%10))
}
