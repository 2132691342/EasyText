package api

import (
	"easy-text/backend/tools"
)

// === 草稿系统 API（🆕 V2.0.0 第一阶段） ===
//
// draftService 在 Startup 阶段 fail-fast 保证非 nil（依赖 DB），
// 此处不再重复 nil 守卫。

// AutoSaveDraft 自动保存草稿
func (h *Handler) AutoSaveDraft(filePath string, content string, encoding string, lineEnding string) error {
	return h.draftService.AutoSave(filePath, content, encoding, lineEnding)
}

// GetDraft 获取单个文件的草稿
func (h *Handler) GetDraft(filePath string) (*tools.DraftEntry, error) {
	return h.draftService.Get(filePath)
}

// ListDrafts 列出所有草稿
func (h *Handler) ListDrafts() ([]tools.DraftEntry, error) {
	return h.draftService.List()
}

// DeleteDraft 删除单个文件的草稿
func (h *Handler) DeleteDraft(filePath string) error {
	return h.draftService.Delete(filePath)
}

// ClearAllDrafts 清除所有草稿
func (h *Handler) ClearAllDrafts() error {
	return h.draftService.ClearAll()
}

// CheckDraftConflict 检查草稿冲突
// 返回: 0=无冲突, 1=草稿更新, 2=磁盘文件更新
func (h *Handler) CheckDraftConflict(filePath string) (int, error) {
	return h.draftService.CheckConflict(filePath)
}
