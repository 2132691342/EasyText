package api

// === Format Conversion ===

// Convert 在支持的格式之间转换（json/yaml/toml/xml）
func (h *Handler) Convert(content, fromFmt, toFmt string) (string, error) {
	return h.convertTool.Convert(content, fromFmt, toFmt)
}
