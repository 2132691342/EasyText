package api

import "easy-text/backend/tools"

// === 宏录制/回放 API ===

// StartMacroRecording 开始录制宏
func (h *Handler) StartMacroRecording() {
	h.macroService.StartRecording()
}

// StopMacroRecording 停止录制并返回保存的宏
func (h *Handler) StopMacroRecording() *tools.Macro {
	return h.macroService.StopRecording()
}

// RecordMacroStep 记录宏步骤
func (h *Handler) RecordMacroStep(step tools.MacroStep) {
	h.macroService.RecordStep(step)
}

// GetMacros 获取所有已保存的宏
func (h *Handler) GetMacros() []tools.Macro {
	return h.macroService.GetMacros()
}

// DeleteMacro 删除指定宏
func (h *Handler) DeleteMacro(id string) bool {
	return h.macroService.DeleteMacro(id)
}

// RenameMacro 重命名宏
func (h *Handler) RenameMacro(id, newName string) bool {
	return h.macroService.RenameMacro(id, newName)
}

// IsMacroRecording 查询是否正在录制
func (h *Handler) IsMacroRecording() bool {
	return h.macroService.IsRecording()
}

// SaveCurrentMacro 保存当前宏（指定名称）
func (h *Handler) SaveCurrentMacro(name string) *tools.Macro {
	return h.macroService.SaveCurrentMacro(name)
}
