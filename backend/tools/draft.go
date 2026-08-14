package tools

import (
	"os"
	"time"

	"easy-text/backend/config"
	"easy-text/backend/utils"

	"gorm.io/gorm"
)

// DraftEntry 草稿条目（API 层使用）
type DraftEntry struct {
	ID          int    `json:"id"`
	FilePath    string `json:"filePath"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
	LineEnding  string `json:"lineEnding"`
	SavedAt     string `json:"savedAt"`     // RFC3339
	FileModtime int64  `json:"fileModtime"` // 磁盘文件修改时间
}

// DraftService 草稿管理服务
type DraftService struct {
	db *gorm.DB
}

// NewDraftService 创建草稿服务
func NewDraftService(db *gorm.DB) *DraftService {
	return &DraftService{db: db}
}

// AutoSave 自动保存草稿
func (s *DraftService) AutoSave(filePath string, content string, encoding string, lineEnding string) error {
	fileModtime := int64(0)
	if info, err := os.Stat(filePath); err == nil {
		fileModtime = info.ModTime().UnixMilli()
	}

	var draft config.Draft
	result := s.db.Where("file_path = ?", filePath).First(&draft)

	if result.Error == gorm.ErrRecordNotFound {
		// 新建草稿
		draft = config.Draft{
			FilePath:    filePath,
			Content:     content,
			Encoding:    encoding,
			LineEnding:  lineEnding,
			FileModtime: fileModtime,
		}
		if err := s.db.Create(&draft).Error; err != nil {
			return utils.ErrDraftSaveFailed
		}
		return nil
	}

	if result.Error != nil {
		return utils.ErrDraftSaveFailed
	}

	// 更新草稿
	draft.Content = content
	draft.Encoding = encoding
	draft.LineEnding = lineEnding
	draft.FileModtime = fileModtime
	draft.SavedAt = time.Now().UnixMilli()

	if err := s.db.Save(&draft).Error; err != nil {
		return utils.ErrDraftSaveFailed
	}

	return nil
}

// Get 获取单个文件的草稿
func (s *DraftService) Get(filePath string) (*DraftEntry, error) {
	var draft config.Draft
	result := s.db.Where("file_path = ?", filePath).First(&draft)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if result.Error != nil {
		return nil, utils.ErrDraftLoadFailed
	}

	return &DraftEntry{
		ID:          draft.ID,
		FilePath:    draft.FilePath,
		Content:     draft.Content,
		Encoding:    draft.Encoding,
		LineEnding:  draft.LineEnding,
		SavedAt:     time.UnixMilli(draft.SavedAt).Format(time.RFC3339),
		FileModtime: draft.FileModtime,
	}, nil
}

// List 列出所有草稿
func (s *DraftService) List() ([]DraftEntry, error) {
	var drafts []config.Draft
	if err := s.db.Order("saved_at DESC").Find(&drafts).Error; err != nil {
		return nil, utils.ErrDraftLoadFailed
	}

	entries := make([]DraftEntry, 0, len(drafts))
	for _, d := range drafts {
		entries = append(entries, DraftEntry{
			ID:          d.ID,
			FilePath:    d.FilePath,
			Content:     d.Content,
			Encoding:    d.Encoding,
			LineEnding:  d.LineEnding,
			SavedAt:     time.UnixMilli(d.SavedAt).Format(time.RFC3339),
			FileModtime: d.FileModtime,
		})
	}
	return entries, nil
}

// Delete 删除单个文件的草稿
func (s *DraftService) Delete(filePath string) error {
	result := s.db.Where("file_path = ?", filePath).Delete(&config.Draft{})
	if result.Error != nil {
		return utils.ErrDraftSaveFailed
	}
	return nil
}

// ClearAll 清除所有草稿
func (s *DraftService) ClearAll() error {
	if err := s.db.Where("1 = 1").Delete(&config.Draft{}).Error; err != nil {
		return utils.ErrDraftSaveFailed
	}
	return nil
}

// CheckConflict 检查草稿与磁盘文件是否冲突
// 返回: 0=无冲突, 1=草稿更新, 2=磁盘文件更新
func (s *DraftService) CheckConflict(filePath string) (int, error) {
	draft, err := s.Get(filePath)
	if err != nil || draft == nil {
		return 0, err
	}

	info, err := os.Stat(filePath)
	if err != nil {
		// 文件已被删除，草稿仍有内容
		return 1, nil
	}

	diskModtime := info.ModTime().UnixMilli()
	if draft.FileModtime < diskModtime {
		// 磁盘文件在草稿保存后又被修改
		return 2, nil
	}
	if draft.FileModtime > diskModtime {
		// 草稿比磁盘文件新（不应出现，但做保护）
		return 1, nil
	}

	return 0, nil
}
