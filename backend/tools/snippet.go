package tools

import (
	"encoding/json"
	"time"

	"easy-text/backend/config"
	"easy-text/backend/utils"

	"gorm.io/gorm"
)

// SnippetEntry 代码片段条目（API 层使用）
type SnippetEntry struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Prefix      string `json:"prefix"`
	Body        string `json:"body"`
	Description string `json:"description"`
	Language    string `json:"language"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// SnippetService 代码片段管理服务
type SnippetService struct {
	db *gorm.DB
}

// NewSnippetService 创建代码片段服务
func NewSnippetService(db *gorm.DB) *SnippetService {
	return &SnippetService{db: db}
}

// GetAll 获取所有片段，可按语言过滤
func (s *SnippetService) GetAll(language string) ([]SnippetEntry, error) {
	var snippets []config.Snippet
	query := s.db.Order("name ASC")
	if language != "" {
		query = query.Where("language = ? OR language = ''", language)
	}
	if err := query.Find(&snippets).Error; err != nil {
		return nil, utils.WrapError(5001, "获取代码片段失败", err)
	}

	entries := make([]SnippetEntry, 0, len(snippets))
	for _, sn := range snippets {
		entries = append(entries, toSnippetEntry(sn))
	}
	return entries, nil
}

// Create 创建代码片段
func (s *SnippetService) Create(entry *SnippetEntry) (int, error) {
	snippet := config.Snippet{
		Name:        entry.Name,
		Prefix:      entry.Prefix,
		Body:        entry.Body,
		Description: entry.Description,
		Language:    entry.Language,
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
	}
	if err := s.db.Create(&snippet).Error; err != nil {
		return 0, utils.WrapError(5001, "创建代码片段失败", err)
	}
	return snippet.ID, nil
}

// Update 更新代码片段
func (s *SnippetService) Update(entry *SnippetEntry) error {
	result := s.db.Model(&config.Snippet{}).Where("id = ?", entry.ID).Updates(map[string]interface{}{
		"name":        entry.Name,
		"prefix":      entry.Prefix,
		"body":        entry.Body,
		"description": entry.Description,
		"language":    entry.Language,
		"updated_at":  time.Now().UnixMilli(),
	})
	if result.Error != nil {
		return utils.WrapError(5001, "更新代码片段失败", result.Error)
	}
	return nil
}

// Delete 删除代码片段
func (s *SnippetService) Delete(id int) error {
	if err := s.db.Delete(&config.Snippet{}, id).Error; err != nil {
		return utils.WrapError(5001, "删除代码片段失败", err)
	}
	return nil
}

// VS Code snippet import format
type vsCodeSnippet struct {
	Prefix      string      `json:"prefix"`
	Body        interface{} `json:"body"` // string or []string
	Description string      `json:"description"`
}

// ImportFromJSON 从 JSON 导入片段（兼容 VS Code 格式）
func (s *SnippetService) ImportFromJSON(jsonData string) (int, error) {
	// 尝试解析为 VS Code 格式（map[string]vsCodeSnippet）
	var vsCodeSnippets map[string]vsCodeSnippet
	if err := json.Unmarshal([]byte(jsonData), &vsCodeSnippets); err == nil && len(vsCodeSnippets) > 0 {
		count := 0
		for name, sn := range vsCodeSnippets {
			body := ""
			switch v := sn.Body.(type) {
			case string:
				body = v
			case []interface{}:
				lines := make([]string, 0, len(v))
				for _, line := range v {
					if s, ok := line.(string); ok {
						lines = append(lines, s)
					}
				}
				// Join lines
				for i, l := range lines {
					if i > 0 {
						body += "\n"
					}
					body += l
				}
			}

			prefix := sn.Prefix
			if prefix == "" {
				prefix = name
			}

			snippet := config.Snippet{
				Name:        name,
				Prefix:      prefix,
				Body:        body,
				Description: sn.Description,
				CreatedAt:   time.Now().UnixMilli(),
				UpdatedAt:   time.Now().UnixMilli(),
			}
			if err := s.db.Create(&snippet).Error; err == nil {
				count++
			}
		}
		return count, nil
	}

	// 尝试解析为 EasyText 格式（[]SnippetEntry）
	var entries []SnippetEntry
	if err := json.Unmarshal([]byte(jsonData), &entries); err != nil {
		return 0, utils.WrapError(2001, "无效的片段导入格式", err)
	}

	count := 0
	for _, entry := range entries {
		snippet := config.Snippet{
			Name:        entry.Name,
			Prefix:      entry.Prefix,
			Body:        entry.Body,
			Description: entry.Description,
			Language:    entry.Language,
			CreatedAt:   time.Now().UnixMilli(),
			UpdatedAt:   time.Now().UnixMilli(),
		}
		if err := s.db.Create(&snippet).Error; err == nil {
			count++
		}
	}
	return count, nil
}

// ExportToJSON 导出所有片段为 JSON
func (s *SnippetService) ExportToJSON() (string, error) {
	var snippets []config.Snippet
	if err := s.db.Order("name ASC").Find(&snippets).Error; err != nil {
		return "", utils.WrapError(5001, "导出代码片段失败", err)
	}

	entries := make([]SnippetEntry, 0, len(snippets))
	for _, sn := range snippets {
		entries = append(entries, toSnippetEntry(sn))
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// toSnippetEntry 将数据库模型转换为 API 模型
func toSnippetEntry(s config.Snippet) SnippetEntry {
	createdAt := ""
	if s.CreatedAt > 0 {
		createdAt = time.UnixMilli(s.CreatedAt).Format(time.RFC3339)
	}
	updatedAt := ""
	if s.UpdatedAt > 0 {
		updatedAt = time.UnixMilli(s.UpdatedAt).Format(time.RFC3339)
	}
	return SnippetEntry{
		ID:          s.ID,
		Name:        s.Name,
		Prefix:      s.Prefix,
		Body:        s.Body,
		Description: s.Description,
		Language:    s.Language,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
