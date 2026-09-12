package tools

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"easy-text/backend/config"
)

// setupDBTest 准备一次性 DB（与 setupRecentTest 同模式）：
// 每个用例独立 TempDir，避免数据串扰；结束时关闭连接让 TempDir 可删。
func setupDBTest(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	if err := config.InitDatabase(filepath.Join(tmpDir, "test.db")); err != nil {
		t.Fatalf("InitDatabase: %v", err)
	}
	return func() { closeDB(config.DB) }
}

// TestDraft_AutoSaveCreateAndUpdate 验证 AutoSave 首次创建、再次调用为更新（同一行记录）。
func TestDraft_AutoSaveCreateAndUpdate(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewDraftService(config.DB)
	path := filepath.Join(t.TempDir(), "note.txt")

	if err := svc.AutoSave(path, "v1", "UTF-8", "CRLF"); err != nil {
		t.Fatalf("AutoSave #1: %v", err)
	}
	if err := svc.AutoSave(path, "v2 content", "UTF-8", "LF"); err != nil {
		t.Fatalf("AutoSave #2: %v", err)
	}

	draft, err := svc.Get(path)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if draft == nil {
		t.Fatal("want draft, got nil")
	}
	if draft.Content != "v2 content" {
		t.Errorf("content mismatch: want %q, got %q", "v2 content", draft.Content)
	}
	if draft.LineEnding != "LF" {
		t.Errorf("lineEnding mismatch: want LF, got %q", draft.LineEnding)
	}
}

// TestDraft_GetMissing 验证不存在的草稿返回 (nil, nil) 而非错误。
func TestDraft_GetMissing(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewDraftService(config.DB)
	draft, err := svc.Get("Z:\\nonexistent\\file.txt")
	if err != nil {
		t.Fatalf("Get(missing): %v", err)
	}
	if draft != nil {
		t.Errorf("want nil for missing draft, got %+v", draft)
	}
}

// TestDraft_ListSortedByTimeDesc 验证 List 按保存时间倒序（最新在前）。
func TestDraft_ListSortedByTimeDesc(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewDraftService(config.DB)
	p1 := filepath.Join(t.TempDir(), "a.txt")
	p2 := filepath.Join(t.TempDir(), "b.txt")

	if err := svc.AutoSave(p1, "older", "UTF-8", "LF"); err != nil {
		t.Fatalf("AutoSave a: %v", err)
	}
	time.Sleep(10 * time.Millisecond) // SavedAt 精度为毫秒
	if err := svc.AutoSave(p2, "newer", "UTF-8", "LF"); err != nil {
		t.Fatalf("AutoSave b: %v", err)
	}

	list, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 drafts, got %d", len(list))
	}
	if list[0].FilePath != p2 {
		t.Errorf("newest first: want %q, got %q", p2, list[0].FilePath)
	}
}

// TestDraft_Delete 验证删除后 Get 返回 nil。
func TestDraft_Delete(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewDraftService(config.DB)
	path := filepath.Join(t.TempDir(), "del.txt")
	if err := svc.AutoSave(path, "x", "UTF-8", "LF"); err != nil {
		t.Fatalf("AutoSave: %v", err)
	}
	if err := svc.Delete(path); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	draft, _ := svc.Get(path)
	if draft != nil {
		t.Errorf("want nil after delete, got %+v", draft)
	}
}

// TestDraft_ClearAll 验证 ClearAll 清空所有草稿。
func TestDraft_ClearAll(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewDraftService(config.DB)
	for i := 0; i < 3; i++ {
		p := filepath.Join(t.TempDir(), "f.txt")
		if err := svc.AutoSave(p, "x", "UTF-8", "LF"); err != nil {
			t.Fatalf("AutoSave #%d: %v", i, err)
		}
	}
	if err := svc.ClearAll(); err != nil {
		t.Fatalf("ClearAll: %v", err)
	}
	list, _ := svc.List()
	if len(list) != 0 {
		t.Errorf("want 0 drafts after clear, got %d", len(list))
	}
}

// TestDraft_CheckConflict 验证冲突判定三分支：
//   - 无草稿 → 0
//   - 磁盘文件在草稿之后被修改 → 2
//   - 文件被删除但草稿仍在 → 1
func TestDraft_CheckConflict(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewDraftService(config.DB)
	dir := t.TempDir()
	path := filepath.Join(dir, "conflict.txt")

	// 无草稿
	code, err := svc.CheckConflict(path)
	if err != nil || code != 0 {
		t.Fatalf("no-draft: want (0, nil), got (%d, %v)", code, err)
	}

	// 有草稿，磁盘未动
	if err := os.WriteFile(path, []byte("disk"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := svc.AutoSave(path, "draft", "UTF-8", "LF"); err != nil {
		t.Fatalf("AutoSave: %v", err)
	}
	code, err = svc.CheckConflict(path)
	if err != nil || code != 0 {
		t.Fatalf("clean: want (0, nil), got (%d, %v)", code, err)
	}

	// 磁盘文件在草稿之后又被修改 → 磁盘更新
	time.Sleep(20 * time.Millisecond)
	if err := os.WriteFile(path, []byte("disk newer"), 0o644); err != nil {
		t.Fatalf("rewrite fixture: %v", err)
	}
	code, err = svc.CheckConflict(path)
	if err != nil || code != 2 {
		t.Fatalf("disk-newer: want (2, nil), got (%d, %v)", code, err)
	}

	// 文件被删除 → 草稿仍在
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove fixture: %v", err)
	}
	code, err = svc.CheckConflict(path)
	if err != nil || code != 1 {
		t.Fatalf("file-removed: want (1, nil), got (%d, %v)", code, err)
	}
}
