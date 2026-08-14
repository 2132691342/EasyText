package api

import "easy-text/backend/tools"

// api_compare.go — 目录对比与二进制对比（薄 handler，委托 tools.CompareService）

// CompareDirectories 对比两个目录，返回按相对路径并集分类的差异条目列表
func (h *Handler) CompareDirectories(leftDir, rightDir string) (*tools.DirCompareResult, error) {
	return h.compareService.CompareDirectories(leftDir, rightDir)
}

// BinaryCompare 逐字节对比两个文件，返回首个差异偏移与 hex 窗口
func (h *Handler) BinaryCompare(leftPath, rightPath string) (*tools.BinCompareResult, error) {
	return h.compareService.BinaryCompare(leftPath, rightPath)
}
