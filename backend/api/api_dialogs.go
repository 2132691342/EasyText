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

// ShowConfirmDialog 显示确认对话框。
//
// ⚠ 平台差异（曾导致线上 Bug）：Windows/Linux 上 MessageDialog 的 Buttons 被框架忽略，
// 返回值只能是标准按钮文本（Yes/No/Ok/Cancel/...），自定义中文按钮文案不生效。
// 因此判定必须基于标准文本集合，不能写成 result == "是"（恒为 false）。
//
// 需要自定义按钮文案、或希望弹窗不夺取窗口焦点时，应改用前端模态层
// （见 frontend/src/utils/confirm.ts），本方法仅作兜底。
func (h *Handler) ShowConfirmDialog(title, message string) (bool, error) {
	result, err := runtime.MessageDialog(h.Ctx, runtime.MessageDialogOptions{
		Title:         title,
		Message:       message,
		Type:          runtime.QuestionDialog,
		DefaultButton: "Yes",
		CancelButton:  "No",
	})
	if err != nil {
		return false, err
	}
	switch result {
	case "Yes", "Ok", "Continue", "Retry", "Try Again", "是", "确定":
		return true, nil
	default:
		return false, nil
	}
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
