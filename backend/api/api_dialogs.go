package api

import "github.com/wailsapp/wails/v2/pkg/runtime"

// === Dialog Helpers ===

// ShowMessageDialog 显示消息对话框
func (h *Handler) ShowMessageDialog(title, message string, dialogType runtime.DialogType) (string, error) {
	return runtime.MessageDialog(h.Ctx, runtime.MessageDialogOptions{
		Title:   title,
		Message: message,
		Type:    dialogType,
	})
}

// ShowConfirmDialog 显示确认对话框
func (h *Handler) ShowConfirmDialog(title, message string) (bool, error) {
	result, err := runtime.MessageDialog(h.Ctx, runtime.MessageDialogOptions{
		Title:         title,
		Message:       message,
		Type:          runtime.QuestionDialog,
		Buttons:       []string{"是", "否"},
		DefaultButton: "是",
		CancelButton:  "否",
	})
	return result == "是", err
}

// === System Info ===

// GetAppVersion 获取应用版本
func (h *Handler) GetAppVersion() string {
	return "1.0.0"
}

// GetSystemInfo 获取系统信息
func (h *Handler) GetSystemInfo() map[string]string {
	return map[string]string{
		"os":      runtime.Environment(h.Ctx).Platform,
		"version": "1.0.0",
	}
}

// Exit 退出应用
func (h *Handler) Exit() {
	runtime.Quit(h.Ctx)
}
