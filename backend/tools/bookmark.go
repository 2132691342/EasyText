package tools

import (
	"time"

	"easy-text/backend/config"
	"easy-text/backend/utils"

	"gorm.io/gorm"
)

// BookmarkEntry 书签条目（API 层使用）
type BookmarkEntry struct {
	ID         int    `json:"id"`
	FilePath   string `json:"filePath"`
	LineNumber int    `json:"lineNumber"`
	Note       string `json:"note"`
	Tag        string `json:"tag"`
	CreatedAt  string `json:"createdAt"` // RFC3339
}

// BookmarkService 书签持久化服务
type BookmarkService struct {
	db *gorm.DB
}

// NewBookmarkService 创建书签服务
func NewBookmarkService(db *gorm.DB) *BookmarkService {
	return &BookmarkService{db: db}
}

// GetByFile 获取指定文件的所有书签
func (s *BookmarkService) GetByFile(filePath string) ([]BookmarkEntry, error) {
	var bookmarks []config.Bookmark
	if err := s.db.Where("file_path = ?", filePath).
		Order("line_number ASC").
		Find(&bookmarks).Error; err != nil {
		return nil, utils.ErrBookmarkLoadFailed
	}

	entries := make([]BookmarkEntry, 0, len(bookmarks))
	for _, b := range bookmarks {
		entries = append(entries, toBookmarkEntry(b))
	}
	return entries, nil
}

// GetAll 获取所有书签，按文件分组
func (s *BookmarkService) GetAll() (map[string][]BookmarkEntry, error) {
	var bookmarks []config.Bookmark
	if err := s.db.Order("file_path ASC, line_number ASC").Find(&bookmarks).Error; err != nil {
		return nil, utils.ErrBookmarkLoadFailed
	}

	result := make(map[string][]BookmarkEntry)
	for _, b := range bookmarks {
		result[b.FilePath] = append(result[b.FilePath], toBookmarkEntry(b))
	}
	return result, nil
}

// Add 添加书签
func (s *BookmarkService) Add(filePath string, lineNumber int, note string, tag string) (*BookmarkEntry, error) {
	// 检查是否已存在（同一行不重复添加）
	var existing config.Bookmark
	result := s.db.Where("file_path = ? AND line_number = ?", filePath, lineNumber).First(&existing)
	if result.Error == nil {
		// 已存在，更新备注
		if note != "" {
			existing.Note = note
		}
		if tag != "" {
			existing.Tag = tag
		}
		s.db.Save(&existing)
		entry := toBookmarkEntry(existing)
		return &entry, nil
	}

	bookmark := config.Bookmark{
		FilePath:   filePath,
		LineNumber: lineNumber,
		Note:       note,
		Tag:        tag,
		CreatedAt:  time.Now().UnixMilli(),
	}
	if err := s.db.Create(&bookmark).Error; err != nil {
		return nil, utils.ErrBookmarkSaveFailed
	}

	entry := toBookmarkEntry(bookmark)
	return &entry, nil
}

// Remove 删除书签
func (s *BookmarkService) Remove(id int) error {
	result := s.db.Delete(&config.Bookmark{}, id)
	if result.Error != nil {
		return utils.ErrBookmarkSaveFailed
	}
	if result.RowsAffected == 0 {
		return utils.ErrBookmarkNotFound
	}
	return nil
}

// UpdateNote 更新书签备注
func (s *BookmarkService) UpdateNote(id int, note string) error {
	result := s.db.Model(&config.Bookmark{}).Where("id = ?", id).Update("note", note)
	if result.Error != nil {
		return utils.ErrBookmarkSaveFailed
	}
	if result.RowsAffected == 0 {
		return utils.ErrBookmarkNotFound
	}
	return nil
}

// UpdateTag 更新书签标签
func (s *BookmarkService) UpdateTag(id int, tag string) error {
	result := s.db.Model(&config.Bookmark{}).Where("id = ?", id).Update("tag", tag)
	if result.Error != nil {
		return utils.ErrBookmarkSaveFailed
	}
	if result.RowsAffected == 0 {
		return utils.ErrBookmarkNotFound
	}
	return nil
}

// toBookmarkEntry 将数据库模型转换为 API 模型
func toBookmarkEntry(b config.Bookmark) BookmarkEntry {
	createdAt := ""
	if b.CreatedAt > 0 {
		createdAt = time.UnixMilli(b.CreatedAt).Format(time.RFC3339)
	}
	return BookmarkEntry{
		ID:         b.ID,
		FilePath:   b.FilePath,
		LineNumber: b.LineNumber,
		Note:       b.Note,
		Tag:        b.Tag,
		CreatedAt:  createdAt,
	}
}
