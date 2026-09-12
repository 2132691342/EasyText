package tools

import (
	"path/filepath"
	"testing"

	"easy-text/backend/config"
)

// TestBookmark_AddAndGetByFile 验证添加书签后按文件查询（按行号升序）。
func TestBookmark_AddAndGetByFile(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewBookmarkService(config.DB)
	file := "C:\\docs\\a.go"

	for _, line := range []int{30, 10, 20} {
		if _, err := svc.Add(file, line, "", ""); err != nil {
			t.Fatalf("Add line %d: %v", line, err)
		}
	}

	list, err := svc.GetByFile(file)
	if err != nil {
		t.Fatalf("GetByFile: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("want 3 bookmarks, got %d", len(list))
	}
	if list[0].LineNumber != 10 || list[1].LineNumber != 20 || list[2].LineNumber != 30 {
		t.Errorf("want ascending [10 20 30], got %v", []int{list[0].LineNumber, list[1].LineNumber, list[2].LineNumber})
	}
}

// TestBookmark_AddDedupSameLine 验证同一行重复添加不产生新记录，且空 note/tag 不覆盖旧值。
func TestBookmark_AddDedupSameLine(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewBookmarkService(config.DB)
	file := "C:\\docs\\b.go"

	first, err := svc.Add(file, 42, "impl here", "todo")
	if err != nil {
		t.Fatalf("Add #1: %v", err)
	}
	second, err := svc.Add(file, 42, "", "")
	if err != nil {
		t.Fatalf("Add #2: %v", err)
	}
	if first.ID != second.ID {
		t.Errorf("want same bookmark ID, got %d vs %d", first.ID, second.ID)
	}

	list, _ := svc.GetByFile(file)
	if len(list) != 1 {
		t.Fatalf("want 1 bookmark after dedup, got %d", len(list))
	}
	if list[0].Note != "impl here" {
		t.Errorf("note overwritten by empty add: want %q, got %q", "impl here", list[0].Note)
	}
}

// TestBookmark_GetAllGrouped 验证 GetAll 按文件分组。
func TestBookmark_GetAllGrouped(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewBookmarkService(config.DB)
	_, err := svc.Add("C:\\a.txt", 1, "", "")
	if err != nil {
		t.Fatalf("Add a1: %v", err)
	}
	_, err = svc.Add("C:\\a.txt", 5, "", "")
	if err != nil {
		t.Fatalf("Add a2: %v", err)
	}
	_, err = svc.Add("C:\\b.txt", 2, "", "")
	if err != nil {
		t.Fatalf("Add b1: %v", err)
	}

	all, err := svc.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("want 2 groups, got %d", len(all))
	}
	if len(all["C:\\a.txt"]) != 2 || len(all["C:\\b.txt"]) != 1 {
		t.Errorf("group sizes: want a=2 b=1, got a=%d b=%d", len(all["C:\\a.txt"]), len(all["C:\\b.txt"]))
	}
}

// TestBookmark_UpdateNoteAndTag 验证 note/tag 更新，且对不存在的 ID 返回错误。
func TestBookmark_UpdateNoteAndTag(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewBookmarkService(config.DB)
	entry, err := svc.Add("C:\\c.go", 7, "", "")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	if err := svc.UpdateNote(entry.ID, "revised"); err != nil {
		t.Fatalf("UpdateNote: %v", err)
	}
	if err := svc.UpdateTag(entry.ID, "bug"); err != nil {
		t.Fatalf("UpdateTag: %v", err)
	}

	list, _ := svc.GetByFile("C:\\c.go")
	if list[0].Note != "revised" || list[0].Tag != "bug" {
		t.Errorf("want note=revised tag=bug, got note=%q tag=%q", list[0].Note, list[0].Tag)
	}

	if err := svc.UpdateNote(99999, "ghost"); err == nil {
		t.Error("want error for nonexistent ID, got nil")
	}
}

// TestBookmark_RemoveAndNotFound 验证删除成功与删除不存在记录的 ErrBookmarkNotFound。
func TestBookmark_RemoveAndNotFound(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewBookmarkService(config.DB)
	entry, err := svc.Add(filepath.Join("C:", "d.go"), 3, "", "")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := svc.Remove(entry.ID); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	list, _ := svc.GetByFile(filepath.Join("C:", "d.go"))
	if len(list) != 0 {
		t.Errorf("want 0 after remove, got %d", len(list))
	}
	if err := svc.Remove(entry.ID); err == nil {
		t.Error("want ErrBookmarkNotFound on double remove, got nil")
	}
}
