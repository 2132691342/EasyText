package file

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// writeLines 写一个包含 n 行（每行 "line<idx>"）的临时文件，返回路径。
func writeLines(t *testing.T, n int) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "partial.txt")
	var sb strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString("line")
		sb.WriteString(strconv.Itoa(i))
	}
	if err := os.WriteFile(p, []byte(sb.String()), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return p
}

// TestReadPartial_BasicRange 验证基本分块读取：offset=2, count=3 返回第 3~5 行。
func TestReadPartial_BasicRange(t *testing.T) {
	p := writeLines(t, 10)
	fr := NewFileReader(0)

	got, err := fr.ReadPartial(p, 2, 3)
	if err != nil {
		t.Fatalf("ReadPartial: %v", err)
	}
	want := "line2\nline3\nline4"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

// TestReadPartial_OffsetBeyondEnd 验证 offset 超出总行数时返回空串不报错。
// 这是前端 loadMoreInBackground 结束条件依赖的语义。
func TestReadPartial_OffsetBeyondEnd(t *testing.T) {
	p := writeLines(t, 5)
	fr := NewFileReader(0)

	got, err := fr.ReadPartial(p, 100, 10)
	if err != nil {
		t.Fatalf("ReadPartial: %v", err)
	}
	if got != "" {
		t.Errorf("offset beyond end should return empty, got %q", got)
	}
}

// TestReadPartial_ZeroCount 验证 count<=0 是 no-op。
func TestReadPartial_ZeroCount(t *testing.T) {
	p := writeLines(t, 5)
	fr := NewFileReader(0)

	got, err := fr.ReadPartial(p, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("count=0 should return empty, got %q", got)
	}
}

// TestReadPartial_PartialTail 验证最后一块不足 count 行时返回剩余所有行。
func TestReadPartial_PartialTail(t *testing.T) {
	p := writeLines(t, 5)
	fr := NewFileReader(0)

	got, err := fr.ReadPartial(p, 3, 10)
	if err != nil {
		t.Fatal(err)
	}
	want := "line3\nline4"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

// TestReadPartial_FileNotFound 验证文件不存在时返回 ErrFileNotFound。
func TestReadPartial_FileNotFound(t *testing.T) {
	fr := NewFileReader(0)
	_, err := fr.ReadPartial(filepath.Join(t.TempDir(), "missing.txt"), 0, 10)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	// AppError 的 Message 应是"文件不存在"
	if !strings.Contains(err.Error(), "文件不存在") {
		t.Errorf("want '文件不存在' in error, got: %v", err)
	}
}
