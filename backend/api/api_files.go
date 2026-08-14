package api

import (
	"os"
	"path/filepath"

	"easy-text/backend/file"
	"easy-text/backend/utils"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// === File Operations ===

// OpenFileDialog 打开文件选择对话框
func (h *Handler) OpenFileDialog() (string, error) {
	result, err := runtime.OpenFileDialog(h.Ctx, runtime.OpenDialogOptions{
		Title: "打开文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "所有文件 (*.*)", Pattern: "*.*"},
			{DisplayName: "文本文件 (*.txt)", Pattern: "*.txt"},
			{DisplayName: "JSON 文件 (*.json)", Pattern: "*.json"},
			{DisplayName: "Markdown 文件 (*.md)", Pattern: "*.md"},
			{DisplayName: "代码文件", Pattern: "*.go;*.js;*.ts;*.py;*.java;*.c;*.cpp;*.h;*.html;*.css;*.xml;*.yaml;*.yml"},
		},
	})
	return result, err
}

// SaveFileDialog 打开保存文件对话框
func (h *Handler) SaveFileDialog(defaultFilename string) (string, error) {
	result, err := runtime.SaveFileDialog(h.Ctx, runtime.SaveDialogOptions{
		Title:           "保存文件",
		DefaultFilename: defaultFilename,
		Filters: []runtime.FileFilter{
			{DisplayName: "所有文件 (*.*)", Pattern: "*.*"},
			{DisplayName: "文本文件 (*.txt)", Pattern: "*.txt"},
			{DisplayName: "JSON 文件 (*.json)", Pattern: "*.json"},
			{DisplayName: "Markdown 文件 (*.md)", Pattern: "*.md"},
		},
	})
	return result, err
}

// ReadFile 读取文件并返回内容
func (h *Handler) ReadFile(path string) (*file.ReadResult, error) {
	result, err := h.fileReader.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ReadPartial 按行分块读取大文件：返回从第 offset 行（0-based）起最多 count 行的文本。
//
// 供前端 LogViewer 的大日志文件分块加载使用——一次性 ReadFile 对 100MB 日志
// 会冻结 UI。详见 file.FileReader.ReadPartial。
func (h *Handler) ReadPartial(path string, offset, count int) (string, error) {
	return h.fileReader.ReadPartial(path, offset, count)
}

// SaveFile 保存文本内容到文件
func (h *Handler) SaveFile(path, content, encoding string) (*file.WriteResult, error) {
	result, err := h.fileWriter.WriteFile(path, content, encoding)
	if err != nil {
		return nil, err
	}

	utils.Log.Info("File saved: %s", path)
	return result, nil
}

// SaveFileWithBackup 带备份保存文件
func (h *Handler) SaveFileWithBackup(path, content, encoding string) (*file.WriteResult, error) {
	return h.fileWriter.WriteFileWithBackup(path, content, encoding)
}

// GetFileInfo 获取文件信息
func (h *Handler) GetFileInfo(path string) (*file.FileInfo, error) {
	return h.fileReader.GetFileInfo(path)
}

// DeleteFile 删除文件
func (h *Handler) DeleteFile(path string) error {
	return file.DeleteFile(path)
}

// DeleteDirectory 删除目录及其所有内容
func (h *Handler) DeleteDirectory(path string) error {
	return file.DeleteDirectory(path)
}

// RenameFile 重命名文件
func (h *Handler) RenameFile(oldPath, newPath string) error {
	return file.RenameFile(oldPath, newPath)
}

// CopyFile 复制文件
func (h *Handler) CopyFile(src, dst string) error {
	return file.CopyFile(src, dst)
}

// IsBinaryFile 检查文件是否为二进制文件
func (h *Handler) IsBinaryFile(path string) (bool, error) {
	return h.fileReader.IsBinaryFile(path)
}

// ReadFileBytes 读取文件字节（用于 PDF、Excel、Word、图片等）
func (h *Handler) ReadFileBytes(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, utils.WrapError(1001, "无法读取文件", err)
	}

	return data, nil
}

// SaveFileBytes 保存字节数据到文件（用于 Excel 等二进制文件）。
// 注意：使用 []int 而非 []byte，因为 Wails v2 将 Go []byte 序列化为 JS []number。
func (h *Handler) SaveFileBytes(path string, data []int) error {
	bytes := make([]byte, len(data))
	for i, v := range data {
		bytes[i] = byte(v)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return utils.WrapError(1002, "无法创建目录", err)
	}

	if err := os.WriteFile(path, bytes, 0644); err != nil {
		return utils.WrapError(1002, "无法写入文件", err)
	}

	utils.Log.Info("Binary file saved: %s", path)
	return nil
}

// StartFileWatch 开始监视指定文件，变更时通过 Wails 事件 "file:change" 通知前端
// fileWatcher 在 NewHandler 阶段 fail-fast 保证非 nil；Ctx 由 Wails 在 Startup 注入。
// 调用本方法应确保 Startup 已执行（Wails 调用顺序保证）。
func (h *Handler) StartFileWatch(path string) error {
	if h.Ctx == nil {
		return utils.NewAppError(1002, "应用上下文未就绪", "")
	}
	return h.fileWatcher.Watch(path, func(evt file.Event) {
		runtime.EventsEmit(h.Ctx, "file:change", evt)
	})
}

// StopFileWatch 停止监视指定文件
func (h *Handler) StopFileWatch(path string) error {
	h.fileWatcher.Unwatch(path)
	return nil
}
