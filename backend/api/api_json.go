package api

import "easy-text/backend/tools"

// === JSON Tools ===

// FormatJSON 格式化 JSON 内容
func (h *Handler) FormatJSON(content string, indentSize int) *tools.JSONResult {
	return h.jsonTool.Format(content, indentSize)
}

// MinifyJSON 压缩 JSON 内容
func (h *Handler) MinifyJSON(content string) *tools.JSONResult {
	return h.jsonTool.Minify(content)
}

// ValidateJSON 校验 JSON 内容
func (h *Handler) ValidateJSON(content string) *tools.JSONResult {
	return h.jsonTool.Validate(content)
}

// FlattenJSON 扁平化嵌套 JSON
func (h *Handler) FlattenJSON(content string, separator string) (string, error) {
	return h.jsonTool.Flatten(content, separator)
}

// ExtractJSONKeys 提取 JSON 所有键
func (h *Handler) ExtractJSONKeys(content string) ([]string, error) {
	return h.jsonTool.ExtractKeys(content)
}

// 🆕 V2.0.0 JSON 工具扩展

// JsonPathQuery 执行 JSONPath 查询
func (h *Handler) JsonPathQuery(jsonStr string, path string) ([]tools.JSONPathResult, error) {
	return h.jsonTool.QueryPath(jsonStr, path)
}

// JsonToStruct 将 JSON 转换为目标语言的结构体定义
func (h *Handler) JsonToStruct(jsonStr string, lang string, rootName string) (string, error) {
	return h.jsonTool.GenerateStruct(jsonStr, lang, rootName)
}

// JsonStructuredDiff 对两个 JSON 进行字段级结构化对比
func (h *Handler) JsonStructuredDiff(leftJSON string, rightJSON string) (*tools.JSONDiffResult, error) {
	return h.jsonTool.StructuredDiff(leftJSON, rightJSON)
}
