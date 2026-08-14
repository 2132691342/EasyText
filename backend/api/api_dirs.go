package api

import (
	"os"
	"path/filepath"
	"strings"

	"easy-text/backend/file"
	"easy-text/backend/utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// === Directory Operations ===

// OpenDirectoryDialog 打开目录选择对话框
func (h *Handler) OpenDirectoryDialog() (string, error) {
	result, err := runtime.OpenDirectoryDialog(h.Ctx, runtime.OpenDialogOptions{
		Title: "打开文件夹",
	})
	return result, err
}

// GetDirectoryTree 获取目录文件树
func (h *Handler) GetDirectoryTree(path string) (*file.FileTree, error) {
	return h.treeBuilder.BuildTree(path)
}

// CreateDirectory 创建新目录
func (h *Handler) CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// ListDirectory 列出目录中的文件
func (h *Handler) ListDirectory(path string) ([]file.FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, utils.WrapError(1002, "无法读取目录", err)
	}

	files := make([]file.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, file.FileInfo{
			Path:     filepath.Join(path, entry.Name()),
			Name:     entry.Name(),
			IsDir:    entry.IsDir(),
			Ext:      strings.ToLower(strings.TrimPrefix(filepath.Ext(entry.Name()), ".")),
			Size:     info.Size(),
			Modified: info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return files, nil
}
