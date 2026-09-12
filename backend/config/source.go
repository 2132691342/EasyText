package config

// Source 提供只读的 AppConfig 视图。业务服务依赖该接口而非全局
// ConfigManager：生产注入适配器，测试可注入 stub。
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
