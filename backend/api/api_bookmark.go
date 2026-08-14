package api

import (
	"easy-text/backend/tools"
)

// === 书签持久化 API（🆕 V2.0.0 第三阶段） ===
//
// bookmarkService 在 Startup 阶段 fail-fast 保证非 nil（依赖 DB），
// 此处不再重复 nil 守卫。

// GetBookmarks 获取指定文件的所有书签
func (h *Handler) GetBookmarks(filePath string) ([]tools.BookmarkEntry, error) {
	return h.bookmarkService.GetByFile(filePath)
}

// GetAllBookmarks 获取所有书签（按文件分组）
func (h *Handler) GetAllBookmarks() (map[string][]tools.BookmarkEntry, error) {
	return h.bookmarkService.GetAll()
}

// AddBookmark 添加书签
func (h *Handler) AddBookmark(filePath string, lineNumber int, note string, tag string) (*tools.BookmarkEntry, error) {
	return h.bookmarkService.Add(filePath, lineNumber, note, tag)
}

// RemoveBookmark 删除书签
func (h *Handler) RemoveBookmark(id int) error {
	return h.bookmarkService.Remove(id)
}

// UpdateBookmarkNote 更新书签备注
func (h *Handler) UpdateBookmarkNote(id int, note string) error {
	return h.bookmarkService.UpdateNote(id, note)
}

// UpdateBookmarkTag 更新书签标签
func (h *Handler) UpdateBookmarkTag(id int, tag string) error {
	return h.bookmarkService.UpdateTag(id, tag)
}
