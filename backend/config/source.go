// Package config 提供应用配置 + 文件持久化 + 单例 ConfigManager。
//
// 本文件（source.go）抽 Source 接口，目的是让业务服务（如 RecentService）
// 通过接口拿配置而**不直接依赖全局单例**。这是 ch11 §DIP（依赖倒置）：
//   - 业务代码持有 Source 接口引用
//   - 生产实现 Source 适配 *ConfigManager
//   - 单元测试可注入 stub/fake 验证业务逻辑而无需读写真实 JSON
package config

// Source 提供只读视图的 AppConfig 给业务服务。
//
// 设计动机：业务代码（RecentService / BookmarkService / SnippetService ...）
// 都在构造期需要 cfg.UI.RecentFilesLimit 等少量字段。原版都直接调
// config.Config.Get()——这是 ch11 §DIP 反模式：测试时无法 stub。
//
// 把 Source 接口单独抽出，限制业务方只能 Get()（不能 Update()），
// 减少业务代码对 ConfigManager 内部的隐式依赖。
type Source interface {
	Get() AppConfig
}

// realSource 是 Source 的生产实现，包装 *ConfigManager。
type realSource struct{ m *ConfigManager }

// NewSource 用 *ConfigManager 构造一个 Source 适配器。
func NewSource(m *ConfigManager) Source {
	if m == nil {
		return NewNoopSource()
	}
	return &realSource{m: m}
}

// Get 实现 Source 接口：返回 AppConfig 副本。
func (s *realSource) Get() AppConfig { return s.m.Get() }

// noopSource 在没有真实 ConfigManager 时返回零值配置（仅 limit 等数值
// 配置会有默认行为）。Production 路径不会用到——Init 失败时已经 panic。
type noopSource struct{}

// NewNoopSource 构造一个总是返回零值 AppConfig 的 Source，仅用于测试。
func NewNoopSource() Source { return &noopSource{} }

// Get 实现 Source 接口。
func (s *noopSource) Get() AppConfig { return AppConfig{} }
