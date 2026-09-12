package tools

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"easy-text/backend/config"
)

// TestDraft_AutoSaveLifecycle 覆盖草稿完整生命周期：AutoSave 首次创建、
// 再次保存为原行更新、List 按保存时间倒序、Delete 单删、ClearAll 清空。
func TestDraft_AutoSaveLifecycle(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewDraftService(config.DB)
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.txt")
	p2 := filepath.Join(dir, "b.txt")

	// 同一路径两次保存应为同一行记录的更新
	if err := svc.AutoSave(p1, "v1", "UTF-8", "CRLF"); err != nil {
		t.Fatalf("AutoSave #1: %v", err)
	}
	if err := svc.AutoSave(p1, "v2 content", "UTF-8", "LF"); err != nil {
		t.Fatalf("AutoSave #2: %v", err)
	}
	draft, err := svc.Get(p1)
	if err != nil || draft == nil {
		t.Fatalf("Get: %v draft=%v", err, draft)
	}
	if draft.Content != "v2 content" || draft.LineEnding != "LF" {
		t.Errorf("draft mismatch: content=%q eol=%q", draft.Content, draft.LineEnding)
	}

	// List 按保存时间倒序（SavedAt 精度为毫秒，需间隔）
	time.Sleep(10 * time.Millisecond)
	if err := svc.AutoSave(p2, "newer", "UTF-8", "LF"); err != nil {
		t.Fatalf("AutoSave p2: %v", err)
	}
	list, err := svc.List()
	if err != nil || len(list) != 2 {
		t.Fatalf("List: err=%v len=%d", err, len(list))
	}
	if list[0].FilePath != p2 {
		t.Errorf("newest first: want %q, got %q", p2, list[0].FilePath)
	}

	// 删除与清空
	if err := svc.Delete(p1); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if draft, _ := svc.Get(p1); draft != nil {
		t.Errorf("want nil after delete, got %+v", draft)
	}
	if err := svc.ClearAll(); err != nil {
		t.Fatalf("ClearAll: %v", err)
	}
	if list, _ := svc.List(); len(list) != 0 {
		t.Errorf("want 0 drafts after clear, got %d", len(list))
	}
}

// TestDraft_CheckConflict 验证冲突判定三分支：
// 无草稿 → 0；磁盘在草稿之后被修改 → 2；文件被删除但草稿仍在 → 1。
func TestDraft_CheckConflict(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewDraftService(config.DB)
	path := filepath.Join(t.TempDir(), "conflict.txt")

	code, err := svc.CheckConflict(path)
	if err != nil || code != 0 {
		t.Fatalf("no-draft: want (0,nil), got (%d,%v)", code, err)
	}

	if err := os.WriteFile(path, []byte("disk"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := svc.AutoSave(path, "draft", "UTF-8", "LF"); err != nil {
		t.Fatalf("AutoSave: %v", err)
	}
	if code, err = svc.CheckConflict(path); err != nil || code != 0 {
		t.Fatalf("clean: want (0,nil), got (%d,%v)", code, err)
	}

	time.Sleep(20 * time.Millisecond)
	if err := os.WriteFile(path, []byte("disk newer"), 0o644); err != nil {
		t.Fatalf("rewrite fixture: %v", err)
	}
	if code, err = svc.CheckConflict(path); err != nil || code != 2 {
		t.Fatalf("disk-newer: want (2,nil), got (%d,%v)", code, err)
	}

	if err := os.Remove(path); err != nil {
		t.Fatalf("remove fixture: %v", err)
	}
	if code, err = svc.CheckConflict(path); err != nil || code != 1 {
		t.Fatalf("file-removed: want (1,nil), got (%d,%v)", code, err)
	}
}
