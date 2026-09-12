package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeFindFixture 写入临时文件并返回路径。
func writeFindFixture(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "find.txt")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return p
}

// TestFindReplace_FileOptions 覆盖单文件查找/替换的全部选项组合语义：
// 大小写敏感开关、全词匹配排除子串、正则模式、非法正则报错、无匹配不写盘。
func TestFindReplace_FileOptions(t *testing.T) {
	svc := NewFindReplaceService()

	t.Run("find-basic-pos", func(t *testing.T) {
		p := writeFindFixture(t, "hello world\nsecond hello here\n")
		matches, err := svc.FindInFile(p, "hello", FindOptions{CaseSensitive: true})
		if err != nil {
			t.Fatalf("FindInFile: %v", err)
		}
		if len(matches) != 2 {
			t.Fatalf("want 2 matches, got %d", len(matches))
		}
		if matches[0].Line != 1 || matches[0].Column != 1 || matches[0].MatchText != "hello" {
			t.Errorf("match[0] wrong: line=%d col=%d text=%q", matches[0].Line, matches[0].Column, matches[0].MatchText)
		}
		if matches[1].Line != 2 || matches[1].Column != 8 {
			t.Errorf("match[1] wrong: line=%d col=%d", matches[1].Line, matches[1].Column)
		}
	})

	t.Run("find-case-insensitive", func(t *testing.T) {
		p := writeFindFixture(t, "Hello World\nHELLO again\n")
		matches, err := svc.FindInFile(p, "hello", FindOptions{CaseSensitive: false})
		if err != nil {
			t.Fatalf("FindInFile: %v", err)
		}
		if len(matches) != 2 {
			t.Errorf("want 2 matches, got %d", len(matches))
		}
	})

	t.Run("find-whole-word", func(t *testing.T) {
		p := writeFindFixture(t, "cat concatenate cat\n")
		matches, err := svc.FindInFile(p, "cat", FindOptions{WholeWord: true, CaseSensitive: true})
		if err != nil {
			t.Fatalf("FindInFile: %v", err)
		}
		if len(matches) != 2 { // concatenate 内部的 cat 必须被排除
			t.Errorf("whole-word want 2 matches, got %d", len(matches))
		}
	})

	t.Run("find-regex", func(t *testing.T) {
		p := writeFindFixture(t, "foo123 bar\nbaz456 foo789\n")
		matches, err := svc.FindInFile(p, `foo\d+`, FindOptions{UseRegex: true})
		if err != nil {
			t.Fatalf("FindInFile: %v", err)
		}
		if len(matches) != 2 || matches[0].MatchText != "foo123" {
			t.Errorf("regex matches wrong: %+v", matches)
		}
	})

	t.Run("find-invalid-regex", func(t *testing.T) {
		p := writeFindFixture(t, "anything\n")
		if _, err := svc.FindInFile(p, `[unclosed`, FindOptions{UseRegex: true}); err == nil {
			t.Error("expected error for invalid regex, got nil")
		}
	})

	t.Run("replace-write-back", func(t *testing.T) {
		p := writeFindFixture(t, "hello world hello\n")
		count, _, err := svc.ReplaceInFile(p, "hello", "hi", FindOptions{CaseSensitive: true})
		if err != nil {
			t.Fatalf("ReplaceInFile: %v", err)
		}
		if count != 2 {
			t.Errorf("want 2 replacements, got %d", count)
		}
		if data, _ := os.ReadFile(p); string(data) != "hi world hi\n" {
			t.Errorf("file content wrong: %q", data)
		}
	})

	t.Run("replace-regex", func(t *testing.T) {
		p := writeFindFixture(t, "a1 b22 c333\n")
		count, _, err := svc.ReplaceInFile(p, `\d+`, "N", FindOptions{UseRegex: true})
		if err != nil {
			t.Fatalf("ReplaceInFile: %v", err)
		}
		if count != 3 {
			t.Errorf("want 3 replacements, got %d", count)
		}
		if data, _ := os.ReadFile(p); string(data) != "aN bN cN\n" {
			t.Errorf("file content wrong: %q", data)
		}
	})

	t.Run("replace-no-match-no-write", func(t *testing.T) {
		original := "unchanged content\n"
		p := writeFindFixture(t, original)
		count, _, err := svc.ReplaceInFile(p, "nothing", "x", FindOptions{})
		if err != nil {
			t.Fatalf("ReplaceInFile: %v", err)
		}
		if count != 0 {
			t.Errorf("want 0, got %d", count)
		}
		if data, _ := os.ReadFile(p); string(data) != original {
			t.Errorf("file must not be rewritten on no-match; got %q", data)
		}
	})
}

// setupFindDir 构造含 n 个匹配文件的临时目录，返回目录路径。
func setupFindDir(t *testing.T, n int) string {
	t.Helper()
	dir := t.TempDir()
	for i := 0; i < n; i++ {
		p := filepath.Join(dir, "file_"+intToStr(i)+".txt")
		if err := os.WriteFile(p, []byte("hello needle world\nsecond line needle here\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}
	return dir
}

// rangeFiles 列出目录下全部 .txt 文件。
func rangeFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		out = append(out, filepath.Join(dir, e.Name()))
	}
	return out
}

// TestFindReplace_DirectoryPipeline 覆盖目录级长链路：百文件并发查找、
// 批量替换写盘、ctx 取消后及时返回不挂死。这是目录搜索功能的核心场景。
func TestFindReplace_DirectoryPipeline(t *testing.T) {
	svc := NewFindReplaceService()

	t.Run("find-in-directory", func(t *testing.T) {
		dir := setupFindDir(t, 100)
		res, err := svc.FindInDirectory(context.Background(), dir, "needle",
			FindOptions{FilePattern: "*.txt", CaseSensitive: false})
		if err != nil {
			t.Fatalf("FindInDirectory: %v", err)
		}
		if len(res) != 100 {
			t.Fatalf("want 100 files matched, got %d", len(res))
		}
		for _, r := range res {
			if r.Count < 1 {
				t.Errorf("file %s: expected ≥1 match, got %d", r.File, r.Count)
			}
		}
	})

	t.Run("batch-replace", func(t *testing.T) {
		dir := setupFindDir(t, 50)
		for _, p := range rangeFiles(t, dir) {
			_ = os.WriteFile(p, []byte("REPLACE_ME is here\n"), 0o644)
		}
		replaced, _, err := svc.BatchReplace(context.Background(), dir, FindOptions{
			Search: "REPLACE_ME", Replace: "DONE", FilePattern: "*.txt",
		})
		if err != nil {
			t.Fatalf("BatchReplace: %v", err)
		}
		if replaced == 0 {
			t.Fatal("want ≥1 replacement, got 0")
		}
	})

	t.Run("ctx-cancel", func(t *testing.T) {
		files := rangeFiles(t, setupFindDir(t, 20))
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		done := make(chan struct{})
		go func() {
			defer close(done)
			_, _ = svc.SearchInFiles(ctx, files, "needle", FindOptions{CaseSensitive: false})
		}()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("SearchInFiles did not return within 2s after ctx cancel")
		}
	})
}
