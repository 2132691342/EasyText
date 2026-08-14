package api

import (
	"context"
	"time"

	"easy-text/backend/tools"
)

// === 查找替换 API（仿 notepad-- findwin 功能）===
//
// 全部方法现在挂在 *SearchHandler 上（通过 Handler 嵌入提升），
// 详见 handler_subsystem.go 的拆分骨架。

// FindInFile 在单个文件中查找
func (h *SearchHandler) FindInFile(filePath, search string, options tools.FindOptions) ([]tools.FindMatch, error) {
	return h.findService.FindInFile(filePath, search, options)
}

// FindInDirectory 在目录中递归查找
func (h *SearchHandler) FindInDirectory(dirPath, search string, options tools.FindOptions) ([]tools.FindInFileResult, error) {
	ctx, cancel := batchCtx()
	defer cancel()
	return h.findService.FindInDirectory(ctx, dirPath, search, options)
}

// ReplaceInFile 在文件中执行替换
func (h *SearchHandler) ReplaceInFile(filePath, search, replace string, options tools.FindOptions) (int, string, error) {
	return h.findService.ReplaceInFile(filePath, search, replace, options)
}

// BatchReplace 批量替换
func (h *SearchHandler) BatchReplace(dirPath string, options tools.FindOptions) (int, []string, error) {
	ctx, cancel := batchCtx()
	defer cancel()
	return h.findService.BatchReplace(ctx, dirPath, options)
}

// 🆕 V2.0.0 全局搜索替换

// SearchInFiles 在指定文件列表中搜索
func (h *SearchHandler) SearchInFiles(filePaths []string, search string, options tools.FindOptions) ([]tools.FindInFileResult, error) {
	ctx, cancel := batchCtx()
	defer cancel()
	return h.findService.SearchInFiles(ctx, filePaths, search, options)
}

// ReplaceInFiles 在指定文件列表中执行替换
func (h *SearchHandler) ReplaceInFiles(filePaths []string, search, replace string, options tools.FindOptions) (int, []string, error) {
	ctx, cancel := batchCtx()
	defer cancel()
	return h.findService.ReplaceInFiles(ctx, filePaths, search, replace, options)
}

// batchCtx 为 Wails API 自动生成的批量操作提供带 30s 超时的 context，
// 替代过去的 context.Background() 让长任务在极端情况下也能兜底收敛（防挂死）。
// 返回 cancel 供调用方 defer 释放，避免 context 资源泄漏（go vet lostcancel）。
func batchCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
