package tools

import (
	"os"
	"path/filepath"
	"time"

	"easy-text/backend/config"
	"easy-text/backend/utils"

	"gorm.io/gorm"
)

// RecentEntryResult 最近访问条目（API 层使用）
type RecentEntryResult struct {
	Path       string `json:"path"`
	IsFolder   bool   `json:"isFolder"`
	Name       string `json:"name"`
	AccessedAt string `json:"accessedAt"` // RFC3339
}

// RecentService 最近访问记录服务
//
// 依赖通过 config.Source 接口注入而非直接读全局单例 config.Config（ch11 DIP）。
// 生产路径由 Handler.Startup 注入 config.NewSource(config.Config)；
// 测试路径可注入 NewNoopSource 或自定义 fake。
type RecentService struct {
	db  *gorm.DB
	cfg config.Source
}

// NewRecentService 创建最近访问记录服务。
//
// cfgSource 不可为 nil；如确实无配置可注入 NewNoopSource()。
func NewRecentService(db *gorm.DB, cfgSource config.Source) *RecentService {
	if cfgSource == nil {
		cfgSource = config.NewNoopSource()
	}
	return &RecentService{db: db, cfg: cfgSource}
}

// defaultRecentLimit 默认最近文件/文件夹保留条数（滚动记录）
const defaultRecentLimit = 10

// Add 添加或更新最近访问记录
func (s *RecentService) Add(path string, isFolder bool) error {
	cfg := s.cfg.Get()
	limit := cfg.UI.RecentFilesLimit
	if limit <= 0 {
		limit = defaultRecentLimit
	}

	name := filepath.Base(path)

	var entry config.RecentEntry
	result := s.db.Where("path = ?", path).First(&entry)

	if result.Error == gorm.ErrRecordNotFound {
		// 新建记录
		entry = config.RecentEntry{
			Path:       path,
			IsFolder:   isFolder,
			Name:       name,
			AccessedAt: time.Now().UnixMilli(),
		}
		if err := s.db.Create(&entry).Error; err != nil {
			return utils.WrapError(5001, "保存最近访问记录失败", err)
		}

		// 清理超出限制的记录
		s.cleanup(isFolder, limit)
		return nil
	}

	if result.Error != nil {
		return utils.WrapError(5001, "保存最近访问记录失败", result.Error)
	}

	// 更新访问时间
	entry.Name = name
	entry.AccessedAt = time.Now().UnixMilli()
	if err := s.db.Save(&entry).Error; err != nil {
		return utils.WrapError(5001, "保存最近访问记录失败", err)
	}

	return nil
}

// GetFiles 获取最近打开的文件
func (s *RecentService) GetFiles() ([]RecentEntryResult, error) {
	cfg := s.cfg.Get()
	limit := cfg.UI.RecentFilesLimit
	if limit <= 0 {
		limit = defaultRecentLimit
	}

	var entries []config.RecentEntry
	if err := s.db.Where("is_folder = ?", false).
		Order("accessed_at DESC").
		Limit(limit).
		Find(&entries).Error; err != nil {
		return nil, utils.WrapError(5001, "获取最近打开文件失败", err)
	}

	results := make([]RecentEntryResult, 0)
	for _, e := range entries {
		// 检查文件是否存在
		if _, err := os.Stat(e.Path); os.IsNotExist(err) {
			// 自动清理不存在的文件
			s.db.Delete(&e)
			continue
		}
		results = append(results, RecentEntryResult{
			Path:       e.Path,
			IsFolder:   e.IsFolder,
			Name:       e.Name,
			AccessedAt: time.UnixMilli(e.AccessedAt).Format(time.RFC3339),
		})
	}
	return results, nil
}

// GetFolders 获取最近打开的文件夹
func (s *RecentService) GetFolders() ([]RecentEntryResult, error) {
	cfg := s.cfg.Get()
	limit := cfg.UI.RecentFilesLimit
	if limit <= 0 {
		limit = defaultRecentLimit
	}

	var entries []config.RecentEntry
	if err := s.db.Where("is_folder = ?", true).
		Order("accessed_at DESC").
		Limit(limit).
		Find(&entries).Error; err != nil {
		return nil, utils.WrapError(5001, "获取最近打开文件夹失败", err)
	}

	results := make([]RecentEntryResult, 0)
	for _, e := range entries {
		if _, err := os.Stat(e.Path); os.IsNotExist(err) {
			s.db.Delete(&e)
			continue
		}
		results = append(results, RecentEntryResult{
			Path:       e.Path,
			IsFolder:   e.IsFolder,
			Name:       e.Name,
			AccessedAt: time.UnixMilli(e.AccessedAt).Format(time.RFC3339),
		})
	}
	return results, nil
}

// ClearFiles 清除所有文件最近访问记录
func (s *RecentService) ClearFiles() error {
	if err := s.db.Where("is_folder = ?", false).Delete(&config.RecentEntry{}).Error; err != nil {
		return utils.WrapError(5001, "清除最近文件记录失败", err)
	}
	return nil
}

// ClearFolders 清除所有文件夹最近访问记录
func (s *RecentService) ClearFolders() error {
	if err := s.db.Where("is_folder = ?", true).Delete(&config.RecentEntry{}).Error; err != nil {
		return utils.WrapError(5001, "清除最近文件夹记录失败", err)
	}
	return nil
}

// cleanup 清理超出限制的记录
func (s *RecentService) cleanup(isFolder bool, limit int) {
	var entries []config.RecentEntry
	s.db.Where("is_folder = ?", isFolder).
		Order("accessed_at DESC").
		Offset(limit).
		Find(&entries)

	for _, e := range entries {
		s.db.Delete(&e)
	}
}
