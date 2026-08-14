package api

import (
	"easy-text/backend/config"
	"easy-text/internal/closepolicy"
)

// === Configuration ===

// GetConfig 获取当前配置
func (h *Handler) GetConfig() config.AppConfig {
	return config.Config.Get()
}

// UpdateConfig 更新配置
func (h *Handler) UpdateConfig(newConfig config.AppConfig) error {
	if err := config.Config.Update(newConfig); err != nil {
		return err
	}
	// 同步关闭策略到 package-global 原子值，下一次窗口关闭立刻生效。
	closepolicy.Set(newConfig.UI.CloseToTray)
	return nil
}

// SetCloseToTray 单项开关：切换"关闭窗口时最小化到托盘"行为。
// 用于前端设置面板无需整体 UpdateConfig 的快速场景，调用后立即同步 closepolicy。
func (h *Handler) SetCloseToTray(enabled bool) error {
	cfg := config.Config.Get()
	cfg.UI.CloseToTray = enabled
	return h.UpdateConfig(cfg)
}

// GetSetting 获取设置值
func (h *Handler) GetSetting(key string) (string, error) {
	return config.GetSetting(key)
}

// SetSetting 设置值
func (h *Handler) SetSetting(key, value string) error {
	return config.SetSetting(key, value)
}

// === Favorites（收藏夹，持久化于 AppConfig.UI.Favorites）===

// AddFavorite 将文件路径加入收藏夹（去重）
func (h *Handler) AddFavorite(path string) error {
	cfg := config.Config.Get()
	for _, f := range cfg.UI.Favorites {
		if f == path {
			return nil
		}
	}
	cfg.UI.Favorites = append(cfg.UI.Favorites, path)
	return config.Config.Update(cfg)
}

// RemoveFavorite 从收藏夹移除文件路径
func (h *Handler) RemoveFavorite(path string) error {
	cfg := config.Config.Get()
	next := make([]string, 0, len(cfg.UI.Favorites))
	for _, f := range cfg.UI.Favorites {
		if f != path {
			next = append(next, f)
		}
	}
	cfg.UI.Favorites = next
	return config.Config.Update(cfg)
}

// GetFavorites 返回收藏夹文件路径列表
func (h *Handler) GetFavorites() []string {
	cfg := config.Config.Get()
	if cfg.UI.Favorites == nil {
		return []string{}
	}
	return cfg.UI.Favorites
}
