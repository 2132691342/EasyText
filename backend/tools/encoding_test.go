package tools

import (
	"bytes"
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// TestEncoding_ConvertRoundTrip 验证 GBK → UTF-8 → GBK 双向往返字节一致，
// 用 x/text 权威编码器生成 GBK fixture，避免手写误字节。
func TestEncoding_ConvertRoundTrip(t *testing.T) {
	et := NewEncodingTool()
	original := "轻量级文本编辑器 EasyText，中文往返测试 123"

	gbkBytes, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(original))
	if err != nil {
		t.Fatalf("GBK encode fixture: %v", err)
	}

	utf8Text, err := et.ToUTF8(gbkBytes, "GBK")
	if err != nil {
		t.Fatalf("ToUTF8(GBK): %v", err)
	}
	if utf8Text != original {
		t.Errorf("ToUTF8 mismatch:\n want %q\n got  %q", original, utf8Text)
	}

	backToGBK, err := et.FromUTF8(utf8Text, "GBK")
	if err != nil {
		t.Fatalf("FromUTF8(GBK): %v", err)
	}
	if !bytes.Equal(backToGBK, gbkBytes) {
		t.Error("GBK round-trip bytes differ")
	}
}

// TestEncoding_BOM 验证 4 族 BOM 识别（含 FF FE 00 00 的 UTF-16LE/UTF-32LE 歧义）
// 以及 AddBOM → HasBOM → RemoveBOM 全编码往返无损；GBK 等无 BOM 编码必须原样透传。
func TestEncoding_BOM(t *testing.T) {
	et := NewEncodingTool()

	cases := []struct {
		name    string
		content []byte
		want    bool
		wantEnc string
	}{
		{"none", []byte("hello"), false, ""},
		{"utf8", []byte{0xEF, 0xBB, 0xBF, 'h', 'i'}, true, "UTF-8"},
		{"utf16le", []byte{0xFF, 0xFE, 'h', 0x00}, true, "UTF-16LE"},
		{"utf16be", []byte{0xFE, 0xFF, 0x00, 'h'}, true, "UTF-16BE"},
		{"utf32le", []byte{0xFF, 0xFE, 0x00, 0x00}, true, "UTF-32LE"},
		{"utf32be", []byte{0x00, 0x00, 0xFE, 0xFF}, true, "UTF-32BE"},
	}
	for _, tc := range cases {
		got, enc := et.HasBOM(tc.content)
		if got != tc.want || enc != tc.wantEnc {
			t.Errorf("%s: want (%v,%q), got (%v,%q)", tc.name, tc.want, tc.wantEnc, got, enc)
		}
	}

	content := []byte("body text")
	for _, enc := range []string{"UTF-8", "UTF-16LE", "UTF-16BE", "UTF-32LE", "UTF-32BE"} {
		withBOM := et.AddBOM(content, enc)
		has, detected := et.HasBOM(withBOM)
		if !has || detected != enc {
			t.Errorf("%s: after AddBOM want (%v,%q), got (%v,%q)", enc, true, enc, has, detected)
		}
		if stripped := et.RemoveBOM(withBOM); !bytes.Equal(stripped, content) {
			t.Errorf("%s: RemoveBOM mismatch: got %q", enc, stripped)
		}
	}

	if got := et.AddBOM(content, "GBK"); !bytes.Equal(got, content) {
		t.Errorf("AddBOM(GBK) should be no-op, got %q", got)
	}
}
