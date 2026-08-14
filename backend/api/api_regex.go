package api

import "easy-text/backend/tools"

// === 正则测试 API（🆕 V2.0.0） ===

// regexTool 正则测试工具实例（延迟初始化）
var regexTool *tools.RegexTool

func (h *Handler) getRegexTool() *tools.RegexTool {
	if regexTool == nil {
		regexTool = tools.NewRegexTool()
	}
	return regexTool
}

// TestRegex 测试正则表达式，返回所有匹配及分组信息
func (h *Handler) TestRegex(pattern string, flags string, input string) *tools.RegexTestResult {
	return h.getRegexTool().TestRegex(pattern, flags, input)
}

// ValidateRegex 验证正则表达式语法合法性
func (h *Handler) ValidateRegex(pattern string) error {
	return h.getRegexTool().ValidateRegex(pattern)
}

// EscapeRegex 转义正则表达式特殊字符
func (h *Handler) EscapeRegex(input string) string {
	return h.getRegexTool().EscapeRegex(input)
}
