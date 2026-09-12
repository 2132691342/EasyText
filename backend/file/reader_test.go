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

// TestReadPartial 覆盖分块读取的核心语义：按 offset/count 取行区间、
// 末块不足时返回剩余行、offset 越界返回空串（前端 loadMore 的停止条件）、
// 文件不存在报"文件不存在"错误。
func TestReadPartial(t *testing.T) {
	fr := NewFileReader(0)

	t.Run("basic-range", func(t *testing.T) {
		got, err := fr.ReadPartial(writeLines(t, 10), 2, 3)
		if err != nil {
			t.Fatalf("ReadPartial: %v", err)
		}
		if want := "line2\nline3\nline4"; got != want {
			t.Errorf("want %q, got %q", want, got)
		}
	})

	t.Run("partial-tail", func(t *testing.T) {
		got, err := fr.ReadPartial(writeLines(t, 5), 3, 10)
		if err != nil {
			t.Fatalf("ReadPartial: %v", err)
		}
		if want := "line3\nline4"; got != want {
			t.Errorf("want %q, got %q", want, got)
		}
	})

	t.Run("offset-beyond-end", func(t *testing.T) {
		got, err := fr.ReadPartial(writeLines(t, 5), 100, 10)
		if err != nil {
			t.Fatalf("ReadPartial: %v", err)
		}
		if got != "" {
			t.Errorf("offset beyond end should return empty, got %q", got)
		}
	})

	t.Run("file-not-found", func(t *testing.T) {
		_, err := fr.ReadPartial(filepath.Join(t.TempDir(), "missing.txt"), 0, 10)
		if err == nil || !strings.Contains(err.Error(), "文件不存在") {
			t.Errorf("want '文件不存在' error, got: %v", err)
		}
	})
}
