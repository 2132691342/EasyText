package tools

import (
	"bytes"
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// TestEncoding_ConvertRoundTripGBK 验证 GBK → UTF-8 → GBK 往返内容一致。
// 用 golang.org/x/text 权威编码器生成 GBK fixture，避免手写误字节。
func TestEncoding_ConvertRoundTripGBK(t *testing.T) {
	et := NewEncodingTool()
	original := "轻量级文本编辑器 EasyText，中文往返测试 123"

	gbkBytes, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(original))
	if err != nil {
		t.Fatalf("GBK encode fixture: %v", err)
	}

	// GBK → UTF-8
	utf8Text, err := et.ToUTF8(gbkBytes, "GBK")
	if err != nil {
		t.Fatalf("ToUTF8(GBK): %v", err)
	}
	if utf8Text != original {
		t.Errorf("ToUTF8 mismatch:\n want %q\n got  %q", original, utf8Text)
	}

	// UTF-8 → GBK
	backToGBK, err := et.FromUTF8(utf8Text, "GBK")
	if err != nil {
		t.Fatalf("FromUTF8(GBK): %v", err)
	}
	if !bytes.Equal(backToGBK, gbkBytes) {
		t.Error("GBK round-trip bytes differ")
	}
}

// TestEncoding_ConvertUnknownEncoding 验证不支持的编码返回错误而非静默失败。
func TestEncoding_ConvertUnknownEncoding(t *testing.T) {
	et := NewEncodingTool()
	if _, err := et.Convert([]byte("hello"), "UTF-8", "NOT-A-REAL-ENCODING"); err == nil {
		t.Error("want error for unknown target encoding, got nil")
	}
	if _, err := et.Convert([]byte("hello"), "NOT-A-REAL-ENCODING", "UTF-8"); err == nil {
		t.Error("want error for unknown source encoding, got nil")
	}
}

// TestEncoding_UTF8Passthrough 验证 UTF-8（getEncoding 返回 nil）直接透传不转码。
func TestEncoding_UTF8Passthrough(t *testing.T) {
	et := NewEncodingTool()
	content := []byte("plain utf-8 text 中文")

	out, err := et.Convert(content, "UTF-8", "UTF-8")
	if err != nil {
		t.Fatalf("Convert UTF-8→UTF-8: %v", err)
	}
	if !bytes.Equal(out, content) {
		t.Errorf("passthrough changed bytes: got %q", out)
	}

	// 空编码名等同 UTF-8
	out2, err := et.Convert(content, "", "")
	if err != nil {
		t.Fatalf("Convert ''→'': %v", err)
	}
	if !bytes.Equal(out2, content) {
		t.Errorf("empty-name passthrough changed bytes: got %q", out2)
	}
}

// TestEncoding_HasBOMAllVariants 验证 4 种 BOM 识别（含 UTF-32LE 与 FF FE 00 00 的歧义判定）。
func TestEncoding_HasBOMAllVariants(t *testing.T) {
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
}

// TestEncoding_AddRemoveBOMRoundTrip 验证 AddBOM → HasBOM → RemoveBOM 往返。
func TestEncoding_AddRemoveBOMRoundTrip(t *testing.T) {
	et := NewEncodingTool()
	content := []byte("body text")

	for _, enc := range []string{"UTF-8", "UTF-16LE", "UTF-16BE", "UTF-32LE", "UTF-32BE"} {
		withBOM := et.AddBOM(content, enc)
		has, detected := et.HasBOM(withBOM)
		if !has {
			t.Errorf("%s: BOM not detected after AddBOM", enc)
		}
		// UTF-16LE 加了 2 字节 BOM 后，若正文以 00 00 开头会被误判为 UTF-32LE；
		// 此处正文为 ASCII，detected 应与添加时一致。
		if detected != enc {
			t.Errorf("%s: detected %q", enc, detected)
		}
		stripped := et.RemoveBOM(withBOM)
		if !bytes.Equal(stripped, content) {
			t.Errorf("%s: RemoveBOM mismatch: got %q", enc, stripped)
		}
	}

	// 未知编码不加 BOM
	if got := et.AddBOM(content, "GBK"); !bytes.Equal(got, content) {
		t.Errorf("AddBOM(GBK) should be no-op, got %q", got)
	}
}

// TestEncoding_SupportedEncodingsList 防回归：编码清单非空且含常用编码。
func TestEncoding_SupportedEncodingsList(t *testing.T) {
	et := NewEncodingTool()
	list := et.SupportedEncodings()
	if len(list) == 0 {
		t.Fatal("want non-empty encodings list")
	}
	nameSet := map[string]bool{}
	for _, e := range list {
		nameSet[e.Name] = true
	}
	for _, want := range []string{"GBK", "Shift_JIS", "UTF-16LE"} {
		if !nameSet[want] {
			t.Errorf("want %q in supported list, missing", want)
		}
	}
}
