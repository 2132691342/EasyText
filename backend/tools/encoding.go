package tools

import (
	"bytes"
	"io"

	"easy-text/backend/file"
	"easy-text/backend/utils"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"
)

// EncodingTool provides encoding conversion functions
type EncodingTool struct{}

// NewEncodingTool creates a new EncodingTool
func NewEncodingTool() *EncodingTool {
	return &EncodingTool{}
}

// SupportedEncodings returns a list of supported encodings
func (et *EncodingTool) SupportedEncodings() []EncodingInfo {
	return []EncodingInfo{
		{Name: "UTF-8", DisplayName: "UTF-8"},
		{Name: "UTF-16LE", DisplayName: "UTF-16 LE"},
		{Name: "UTF-16BE", DisplayName: "UTF-16 BE"},
		{Name: "UTF-32LE", DisplayName: "UTF-32 LE"},
		{Name: "UTF-32BE", DisplayName: "UTF-32 BE"},
		{Name: "GB18030", DisplayName: "简体中文 (GB18030)"},
		{Name: "GBK", DisplayName: "简体中文 (GBK)"},
		{Name: "Big5", DisplayName: "繁体中文 (Big5)"},
		{Name: "ISO-8859-1", DisplayName: "西欧语言 (ISO-8859-1)"},
		{Name: "ISO-8859-2", DisplayName: "中欧语言 (ISO-8859-2)"},
		{Name: "ISO-8859-3", DisplayName: "南欧语言 (ISO-8859-3)"},
		{Name: "ISO-8859-4", DisplayName: "北欧语言 (ISO-8859-4)"},
		{Name: "ISO-8859-5", DisplayName: "西里尔字母 (ISO-8859-5)"},
		{Name: "ISO-8859-6", DisplayName: "阿拉伯语 (ISO-8859-6)"},
		{Name: "ISO-8859-7", DisplayName: "希腊语 (ISO-8859-7)"},
		{Name: "ISO-8859-8", DisplayName: "希伯来语 (ISO-8859-8)"},
		{Name: "ISO-8859-9", DisplayName: "土耳其语 (ISO-8859-9)"},
		{Name: "ISO-8859-10", DisplayName: "北欧语言 (ISO-8859-10)"},
		{Name: "ISO-8859-13", DisplayName: "波罗的海语言 (ISO-8859-13)"},
		{Name: "ISO-8859-14", DisplayName: "凯尔特语 (ISO-8859-14)"},
		{Name: "ISO-8859-15", DisplayName: "西欧语言 (ISO-8859-15)"},
		{Name: "ISO-8859-16", DisplayName: "东南欧语言 (ISO-8859-16)"},
		{Name: "Windows-1250", DisplayName: "中欧语言 (Windows-1250)"},
		{Name: "Windows-1251", DisplayName: "西里尔字母 (Windows-1251)"},
		{Name: "Windows-1252", DisplayName: "西欧语言 (Windows-1252)"},
		{Name: "Windows-1253", DisplayName: "希腊语 (Windows-1253)"},
		{Name: "Windows-1254", DisplayName: "土耳其语 (Windows-1254)"},
		{Name: "Windows-1255", DisplayName: "希伯来语 (Windows-1255)"},
		{Name: "Windows-1256", DisplayName: "阿拉伯语 (Windows-1256)"},
		{Name: "Windows-1257", DisplayName: "波罗的海语言 (Windows-1257)"},
		{Name: "Windows-1258", DisplayName: "越南语 (Windows-1258)"},
		{Name: "Shift_JIS", DisplayName: "日语 (Shift_JIS)"},
		{Name: "EUC-JP", DisplayName: "日语 (EUC-JP)"},
		{Name: "ISO-2022-JP", DisplayName: "日语 (ISO-2022-JP)"},
		{Name: "EUC-KR", DisplayName: "韩语 (EUC-KR)"},
		{Name: "ISO-2022-KR", DisplayName: "韩语 (ISO-2022-KR)"},
		{Name: "KOI8-R", DisplayName: "俄语 (KOI8-R)"},
		{Name: "KOI8-U", DisplayName: "乌克兰语 (KOI8-U)"},
	}
}

// EncodingInfo represents encoding information
type EncodingInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// Convert converts text from one encoding to another
func (et *EncodingTool) Convert(content []byte, fromEncoding, toEncoding string) ([]byte, error) {
	// Get source encoding
	srcEnc, err := et.getEncoding(fromEncoding)
	if err != nil {
		return nil, err
	}

	// Get target encoding
	dstEnc, err := et.getEncoding(toEncoding)
	if err != nil {
		return nil, err
	}

	// Decode from source encoding
	var decoded []byte
	if srcEnc != nil {
		decoder := srcEnc.NewDecoder()
		decoded, err = io.ReadAll(transform.NewReader(bytes.NewReader(content), decoder))
		if err != nil {
			return nil, utils.WrapError(3001, "解码失败", err)
		}
	} else {
		decoded = content
	}

	// Encode to target encoding
	if dstEnc != nil {
		encoder := dstEnc.NewEncoder()
		encoded, err := io.ReadAll(transform.NewReader(bytes.NewReader(decoded), encoder))
		if err != nil {
			return nil, utils.WrapError(3001, "编码失败", err)
		}
		return encoded, nil
	}

	return decoded, nil
}

// ToUTF8 converts text to UTF-8
func (et *EncodingTool) ToUTF8(content []byte, fromEncoding string) (string, error) {
	enc, err := et.getEncoding(fromEncoding)
	if err != nil {
		return "", err
	}

	if enc == nil {
		return string(content), nil
	}

	decoder := enc.NewDecoder()
	decoded, err := io.ReadAll(transform.NewReader(bytes.NewReader(content), decoder))
	if err != nil {
		return "", utils.WrapError(3001, "解码失败", err)
	}

	return string(decoded), nil
}

// FromUTF8 converts UTF-8 text to specified encoding
func (et *EncodingTool) FromUTF8(content string, toEncoding string) ([]byte, error) {
	enc, err := et.getEncoding(toEncoding)
	if err != nil {
		return nil, err
	}

	if enc == nil {
		return []byte(content), nil
	}

	encoder := enc.NewEncoder()
	encoded, err := io.ReadAll(transform.NewReader(bytes.NewReader([]byte(content)), encoder))
	if err != nil {
		return nil, utils.WrapError(3001, "编码失败", err)
	}

	return encoded, nil
}

// getEncoding gets an encoding by name
func (et *EncodingTool) getEncoding(name string) (encoding.Encoding, error) {
	if name == "" || name == "UTF-8" {
		return nil, nil
	}

	enc, err := ianaindex.IANA.Encoding(name)
	if err != nil || enc == nil {
		return nil, utils.ErrUnsupportedEncoding
	}
	return enc, nil
}

// DetectEncoding detects the encoding of content.
//
// Deprecated: 该方法仅做薄包装，请直接调用 file.DetectEncoding 避免一次间接调用。
// 保留此方法仅为兼容现有调用方，后续 API 整理可删除。
func (et *EncodingTool) DetectEncoding(content []byte) string {
	name, _ := file.DetectEncoding(content)
	return name
}

// isLikelyGBK checks if content is likely GBK encoded
func (et *EncodingTool) isLikelyGBK(content []byte) bool {
	// Count potential GBK characters
	gbkCount := 0
	totalCount := 0

	for i := 0; i < len(content); i++ {
		if content[i] >= 0x81 && content[i] <= 0xFE {
			if i+1 < len(content) {
				secondByte := content[i+1]
				if (secondByte >= 0x40 && secondByte <= 0x7E) || (secondByte >= 0x80 && secondByte <= 0xFE) {
					gbkCount++
					i++
				}
			}
			totalCount++
		}
	}

	// If more than 10% of potential double-byte characters are valid GBK, likely GBK
	if totalCount > 0 && float64(gbkCount)/float64(totalCount) > 0.1 {
		return true
	}
	return false
}

// HasBOM checks if content has a BOM
func (et *EncodingTool) HasBOM(content []byte) (bool, string) {
	if len(content) >= 3 {
		if content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
			return true, "UTF-8"
		}
	}
	if len(content) >= 2 {
		if content[0] == 0xFF && content[1] == 0xFE {
			if len(content) >= 4 && content[2] == 0x00 && content[3] == 0x00 {
				return true, "UTF-32LE"
			}
			return true, "UTF-16LE"
		}
		if content[0] == 0xFE && content[1] == 0xFF {
			return true, "UTF-16BE"
		}
	}
	if len(content) >= 4 {
		if content[0] == 0x00 && content[1] == 0x00 && content[2] == 0xFE && content[3] == 0xFF {
			return true, "UTF-32BE"
		}
	}
	return false, ""
}

// RemoveBOM removes BOM from content
func (et *EncodingTool) RemoveBOM(content []byte) []byte {
	if len(content) >= 3 {
		if content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
			return content[3:]
		}
	}
	if len(content) >= 2 {
		if content[0] == 0xFF && content[1] == 0xFE {
			if len(content) >= 4 && content[2] == 0x00 && content[3] == 0x00 {
				return content[4:]
			}
			return content[2:]
		}
		if content[0] == 0xFE && content[1] == 0xFF {
			return content[2:]
		}
	}
	return content
}

// AddBOM adds BOM to content
func (et *EncodingTool) AddBOM(content []byte, encoding string) []byte {
	switch encoding {
	case "UTF-8":
		return append([]byte{0xEF, 0xBB, 0xBF}, content...)
	case "UTF-16LE":
		return append([]byte{0xFF, 0xFE}, content...)
	case "UTF-16BE":
		return append([]byte{0xFE, 0xFF}, content...)
	case "UTF-32LE":
		return append([]byte{0xFF, 0xFE, 0x00, 0x00}, content...)
	case "UTF-32BE":
		return append([]byte{0x00, 0x00, 0xFE, 0xFF}, content...)
	default:
		return content
	}
}
