package api

import (
	"easy-text/backend/tools"
)

// === 代码片段 API（🆕 V2.0.0 第二阶段） ===
//
// snippetService 在 Startup 阶段 fail-fast 保证非 nil（依赖 DB），
// 此处不再重复 nil 守卫。

// GetSnippets 获取代码片段列表，可按语言过滤
func (h *Handler) GetSnippets(language string) ([]tools.SnippetEntry, error) {
	return h.snippetService.GetAll(language)
}

// CreateSnippet 创建代码片段
func (h *Handler) CreateSnippet(entry tools.SnippetEntry) (int, error) {
	return h.snippetService.Create(&entry)
}

// UpdateSnippet 更新代码片段
func (h *Handler) UpdateSnippet(entry tools.SnippetEntry) error {
	return h.snippetService.Update(&entry)
}

// DeleteSnippet 删除代码片段
func (h *Handler) DeleteSnippet(id int) error {
	return h.snippetService.Delete(id)
}

// ImportSnippets 导入代码片段（兼容 VS Code 格式）
func (h *Handler) ImportSnippets(jsonData string) (int, error) {
	return h.snippetService.ImportFromJSON(jsonData)
}

// ExportSnippets 导出所有代码片段为 JSON
func (h *Handler) ExportSnippets() (string, error) {
	return h.snippetService.ExportToJSON()
}
