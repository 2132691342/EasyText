package config

import "testing"

// TestDefaultConfig_CloseToTrayTrue 验证默认配置开启托盘驻留：与用户提问意图一致，
// 修复「默认行为应为最小化到托盘」的回归基线。
func TestDefaultConfig_CloseToTrayTrue(t *testing.T) {
	if !defaultConfig.UI.CloseToTray {
		t.Error("default UI.CloseToTray should be true (minimize-to-tray by default)")
	}
}

// TestDefaultConfig_RecentFilesLimitIs10 防御：默认上限 10，避免被误改回 20。
func TestDefaultConfig_RecentFilesLimitIs10(t *testing.T) {
	if defaultConfig.UI.RecentFilesLimit != 10 {
		t.Errorf("default RecentFilesLimit: want 10, got %d", defaultConfig.UI.RecentFilesLimit)
	}
}

// TestMergeConfig_V3PreservesCloseToTray 验证 V3 schema 合并时 CloseToTray 字段被保留：
// 这是切换"关闭直接退出"后保存、再读取时不应被默认值覆盖。
func TestMergeConfig_V3PreservesCloseToTray(t *testing.T) {
	loaded := defaultConfig // 复制
	loaded.Version = 3
	loaded.UI.CloseToTray = false // 用户主动关掉了托盘

	merged := mergeConfig(defaultConfig, loaded)
	if merged.UI.CloseToTray {
		t.Error("V3 schema CloseToTray should not be clobbered to default")
	}
}

// TestMergeConfig_V2MigratesCloseToTray 验证旧 V2 配置文件（JSON 缺失 CloseToTray 字段）
// 升级到 V3 时会落到默认值（true）。
func TestMergeConfig_V2MigratesCloseToTray(t *testing.T) {
	loaded := defaultConfig
	loaded.Version = 2
	loaded.UI.CloseToTray = false // V2 schema 字段会被忽略；仅用于让结构对齐

	merged := mergeConfig(defaultConfig, loaded)
	if !merged.UI.CloseToTray {
		t.Error("V2→V3 migration should fall back to default CloseToTray=true")
	}
	if merged.Version != 3 {
		t.Errorf("merged version should bump to 3, got %d", merged.Version)
	}
}
