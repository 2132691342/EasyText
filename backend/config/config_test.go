package config

import "testing"

// TestConfig_MergeAndMigration 覆盖配置合并语义：默认值基线（托盘驻留开启、
// 最近文件上限 10）、V3 schema 用户设置不被默认值覆盖、V2 旧配置升级时
// 新字段回落默认值且版本号升到 3。
func TestConfig_MergeAndMigration(t *testing.T) {
	if !defaultConfig.UI.CloseToTray {
		t.Error("default UI.CloseToTray should be true")
	}
	if defaultConfig.UI.RecentFilesLimit != 10 {
		t.Errorf("default RecentFilesLimit: want 10, got %d", defaultConfig.UI.RecentFilesLimit)
	}

	// V3：用户主动关闭托盘驻留后保存，重读时不得被默认值覆盖
	v3 := defaultConfig
	v3.Version = 3
	v3.UI.CloseToTray = false
	if merged := mergeConfig(defaultConfig, v3); merged.UI.CloseToTray {
		t.Error("V3 CloseToTray should not be clobbered to default")
	}

	// V2：旧配置缺 CloseToTray 字段，升级到 V3 时回落默认值
	v2 := defaultConfig
	v2.Version = 2
	v2.UI.CloseToTray = false
	merged := mergeConfig(defaultConfig, v2)
	if !merged.UI.CloseToTray {
		t.Error("V2→V3 migration should fall back to default CloseToTray=true")
	}
	if merged.Version != 3 {
		t.Errorf("merged version should bump to 3, got %d", merged.Version)
	}
}
