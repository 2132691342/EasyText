package api

import "easy-text/backend/tools"

// === Diff Tools ===
//
// 全部方法现在挂在 *SearchHandler 上（通过 Handler 嵌入提升），
// 详见 handler_subsystem.go 的拆分骨架。

// CompareDiff 比较两个文本并返回差异
func (h *SearchHandler) CompareDiff(oldText, newText string) (*tools.DiffResult, error) {
	return h.diffTool.Compare(oldText, newText)
}

// CompareDiffLines 逐行比较两个文本
func (h *SearchHandler) CompareDiffLines(oldText, newText string) ([]tools.DiffBlock, error) {
	return h.diffTool.CompareLines(oldText, newText)
}

// GetDiffPatch 生成 diff patch
func (h *SearchHandler) GetDiffPatch(oldText, newText string) string {
	return h.diffTool.GetPatch(oldText, newText)
}

// ApplyDiffPatch 应用 diff patch
func (h *SearchHandler) ApplyDiffPatch(text, patch string) (string, error) {
	return h.diffTool.ApplyPatch(text, patch)
}

// 🆕 V2.0.0 CompareCharacters 字符级差异比较，返回 HTML 格式（<ins>/<del> 标记）
func (h *SearchHandler) CompareCharacters(oldStr, newStr string) string {
	return h.diffTool.CompareCharacters(oldStr, newStr)
}
