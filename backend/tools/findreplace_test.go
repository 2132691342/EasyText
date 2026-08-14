package tools

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFindFixture 写一个临时文件，返回路径与内容。
func writeFindFixture(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "find.txt")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return p
}

// TestFindInFile_Basic 验证普通大小写敏感匹配：命中行号 / 列号 / 匹配文本。
func TestFindInFile_Basic(t *testing.T) {
	p := writeFindFixture(t, "hello world\nsecond hello here\nthird line\n")
	svc := NewFindReplaceService()

	matches, err := svc.FindInFile(p, "hello", FindOptions{CaseSensitive: true})
	if err != nil {
		t.Fatalf("FindInFile: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("want 2 matches, got %d", len(matches))
	}
	// 第一处：第 1 行，第 1 列
	if matches[0].Line != 1 || matches[0].Column != 1 || matches[0].MatchText != "hello" {
		t.Errorf("match[0] wrong: line=%d col=%d text=%q", matches[0].Line, matches[0].Column, matches[0].MatchText)
	}
	// 第二处：第 2 行，第 8 列（"second " 占 7 字节，hello 从第 8 列开始）
	if matches[1].Line != 2 || matches[1].Column != 8 {
		t.Errorf("match[1] wrong: line=%d col=%d", matches[1].Line, matches[1].Column)
	}
}

// TestFindInFile_CaseInsensitive 验证大小写不敏感匹配。
func TestFindInFile_CaseInsensitive(t *testing.T) {
	p := writeFindFixture(t, "Hello World\nHELLO again\n")
	svc := NewFindReplaceService()

	matches, err := svc.FindInFile(p, "hello", FindOptions{CaseSensitive: false})
	if err != nil {
		t.Fatalf("FindInFile: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("case-insensitive want 2 matches, got %d", len(matches))
	}
}

// TestFindInFile_WholeWord 验证全词匹配：排除子串命中。
func TestFindInFile_WholeWord(t *testing.T) {
	p := writeFindFixture(t, "cat concatenate cat\n")
	svc := NewFindReplaceService()

	matches, err := svc.FindInFile(p, "cat", FindOptions{WholeWord: true, CaseSensitive: true})
	if err != nil {
		t.Fatalf("FindInFile: %v", err)
	}
	// "concatenate" 里的 "cat" 不是独立单词，应被排除；只有两个独立的 "cat" 命中
	if len(matches) != 2 {
		t.Fatalf("whole-word want 2 matches, got %d", len(matches))
	}
}

// TestFindInFile_Regex 验证正则模式。
func TestFindInFile_Regex(t *testing.T) {
	p := writeFindFixture(t, "foo123 bar\nbaz456 foo789\n")
	svc := NewFindReplaceService()

	matches, err := svc.FindInFile(p, `foo\d+`, FindOptions{UseRegex: true})
	if err != nil {
		t.Fatalf("FindInFile regex: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("regex want 2 matches, got %d", len(matches))
	}
	if matches[0].MatchText != "foo123" {
		t.Errorf("match[0] = %q, want foo123", matches[0].MatchText)
	}
}

// TestFindInFile_NoMatch 验证无匹配返回空切片（非 nil 语义留给调用方判断）。
func TestFindInFile_NoMatch(t *testing.T) {
	p := writeFindFixture(t, "abc def\n")
	svc := NewFindReplaceService()

	matches, err := svc.FindInFile(p, "zzz", FindOptions{})
	if err != nil {
		t.Fatalf("FindInFile: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("want 0 matches, got %d", len(matches))
	}
}

// TestFindInFile_InvalidRegex 验证非法正则返回错误而非 panic。
func TestFindInFile_InvalidRegex(t *testing.T) {
	p := writeFindFixture(t, "anything\n")
	svc := NewFindReplaceService()

	_, err := svc.FindInFile(p, `[unclosed`, FindOptions{UseRegex: true})
	if err == nil {
		t.Fatal("expected error for invalid regex, got nil")
	}
}

// TestReplaceInFile_Basic 验证普通替换写入文件并返回替换次数。
func TestReplaceInFile_Basic(t *testing.T) {
	p := writeFindFixture(t, "hello world hello\n")
	svc := NewFindReplaceService()

	count, _, err := svc.ReplaceInFile(p, "hello", "hi", FindOptions{CaseSensitive: true})
	if err != nil {
		t.Fatalf("ReplaceInFile: %v", err)
	}
	if count != 2 {
		t.Errorf("want 2 replacements, got %d", count)
	}

	// 读回验证
	data, _ := os.ReadFile(p)
	if got := string(data); got != "hi world hi\n" {
		t.Errorf("file content wrong: %q", got)
	}
}

// TestReplaceInFile_CaseInsensitive 验证不区分大小写替换。
func TestReplaceInFile_CaseInsensitive(t *testing.T) {
	p := writeFindFixture(t, "Hello HELLO\n")
	svc := NewFindReplaceService()

	count, _, err := svc.ReplaceInFile(p, "hello", "hi", FindOptions{CaseSensitive: false})
	if err != nil {
		t.Fatalf("ReplaceInFile: %v", err)
	}
	if count != 2 {
		t.Errorf("want 2 replacements, got %d", count)
	}
	data, _ := os.ReadFile(p)
	if got := string(data); got != "hi hi\n" {
		t.Errorf("file content wrong: %q", got)
	}
}

// TestReplaceInFile_Regex 验证正则替换。
func TestReplaceInFile_Regex(t *testing.T) {
	p := writeFindFixture(t, "a1 b22 c333\n")
	svc := NewFindReplaceService()

	count, _, err := svc.ReplaceInFile(p, `\d+`, "N", FindOptions{UseRegex: true})
	if err != nil {
		t.Fatalf("ReplaceInFile regex: %v", err)
	}
	if count != 3 {
		t.Errorf("want 3 replacements, got %d", count)
	}
	data, _ := os.ReadFile(p)
	if got := string(data); got != "aN bN cN\n" {
		t.Errorf("file content wrong: %q", got)
	}
}

// TestReplaceInFile_NoMatch_NoWrite 验证无匹配时不写文件（count=0 且文件不变）。
func TestReplaceInFile_NoMatch_NoWrite(t *testing.T) {
	original := "unchanged content\n"
	p := writeFindFixture(t, original)
	svc := NewFindReplaceService()

	count, _, err := svc.ReplaceInFile(p, "nothing", "x", FindOptions{})
	if err != nil {
		t.Fatalf("ReplaceInFile: %v", err)
	}
	if count != 0 {
		t.Errorf("want 0, got %d", count)
	}
	data, _ := os.ReadFile(p)
	if got := string(data); got != original {
		t.Errorf("file should not be rewritten on no-match; got %q", got)
	}
}
