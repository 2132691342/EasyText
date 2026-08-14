package api

import "easy-text/backend/tools"

// === Encoding Tools ===

// GetSupportedEncodings 获取支持的编码列表
func (h *Handler) GetSupportedEncodings() []tools.EncodingInfo {
	return h.encodingTool.SupportedEncodings()
}

// ConvertEncoding 编码转换
func (h *Handler) ConvertEncoding(content []byte, fromEncoding, toEncoding string) ([]byte, error) {
	return h.encodingTool.Convert(content, fromEncoding, toEncoding)
}

// DetectEncoding 检测编码
func (h *Handler) DetectEncoding(content []byte) string {
	return h.encodingTool.DetectEncoding(content)
}

// ConvertToUTF8 转换为 UTF-8
func (h *Handler) ConvertToUTF8(content []byte, fromEncoding string) (string, error) {
	return h.encodingTool.ToUTF8(content, fromEncoding)
}

// ConvertFromUTF8 从 UTF-8 转换为指定编码
func (h *Handler) ConvertFromUTF8(content string, toEncoding string) ([]byte, error) {
	return h.encodingTool.FromUTF8(content, toEncoding)
}

// HasBOM 检测内容是否包含 BOM，返回 (hasBOM, encodingName)
func (h *Handler) HasBOM(content []byte) (bool, string) {
	return h.encodingTool.HasBOM(content)
}

// RemoveBOM 移除 BOM
func (h *Handler) RemoveBOM(content []byte) []byte {
	return h.encodingTool.RemoveBOM(content)
}

// AddBOM 添加 BOM
func (h *Handler) AddBOM(content []byte, encoding string) []byte {
	return h.encodingTool.AddBOM(content, encoding)
}
