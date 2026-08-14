package api

import (
	"easy-text/backend/tools"
)

// === 脚本扩展 API（🆕 V2.0.0 第四阶段） ===
//
// scriptService 不依赖 DB，在 Startup 阶段始终构造，恒非 nil。

// ListScripts 列出所有脚本
func (h *Handler) ListScripts() ([]tools.ScriptInfo, error) {
	return h.scriptService.List()
}

// GetScript 获取单个脚本
func (h *Handler) GetScript(id string) (*tools.ScriptInfo, error) {
	return h.scriptService.Get(id)
}

// SaveScript 保存脚本
func (h *Handler) SaveScript(script tools.ScriptInfo) error {
	return h.scriptService.Save(script)
}

// DeleteScript 删除脚本
func (h *Handler) DeleteScript(id string) error {
	return h.scriptService.Delete(id)
}

// ExecuteScript 执行脚本
func (h *Handler) ExecuteScript(id string, context tools.ScriptContext) (*tools.ScriptResult, error) {
	return h.scriptService.Execute(id, context)
}
