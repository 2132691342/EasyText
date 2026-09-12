package tools

import (
	"testing"

	"easy-text/backend/config"
)

// TestBookmark_Lifecycle 覆盖书签完整生命周期：乱序添加后按行号升序返回、
// GetAll 按文件分组、note/tag 更新（不存在 ID 报错）、删除（重复删除报错）。
func TestBookmark_Lifecycle(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewBookmarkService(config.DB)
	fileA, fileB := "C:\\docs\\a.go", "C:\\docs\\b.go"

	for _, line := range []int{30, 10, 20} {
		if _, err := svc.Add(fileA, line, "", ""); err != nil {
			t.Fatalf("Add line %d: %v", line, err)
		}
	}
	entryB, err := svc.Add(fileB, 2, "", "")
	if err != nil {
		t.Fatalf("Add b: %v", err)
	}

	list, err := svc.GetByFile(fileA)
	if err != nil || len(list) != 3 {
		t.Fatalf("GetByFile: err=%v len=%d", err, len(list))
	}
	if list[0].LineNumber != 10 || list[1].LineNumber != 20 || list[2].LineNumber != 30 {
		t.Errorf("want ascending [10 20 30], got %v", []int{list[0].LineNumber, list[1].LineNumber, list[2].LineNumber})
	}

	all, err := svc.GetAll()
	if err != nil || len(all) != 2 {
		t.Fatalf("GetAll: err=%v groups=%d", err, len(all))
	}
	if len(all[fileA]) != 3 || len(all[fileB]) != 1 {
		t.Errorf("group sizes: want a=3 b=1, got a=%d b=%d", len(all[fileA]), len(all[fileB]))
	}

	if err := svc.UpdateNote(list[0].ID, "revised"); err != nil {
		t.Fatalf("UpdateNote: %v", err)
	}
	if err := svc.UpdateTag(list[0].ID, "bug"); err != nil {
		t.Fatalf("UpdateTag: %v", err)
	}
	list, _ = svc.GetByFile(fileA)
	if list[0].Note != "revised" || list[0].Tag != "bug" {
		t.Errorf("want note=revised tag=bug, got note=%q tag=%q", list[0].Note, list[0].Tag)
	}
	if err := svc.UpdateNote(99999, "ghost"); err == nil {
		t.Error("want error for nonexistent ID, got nil")
	}

	if err := svc.Remove(entryB.ID); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if err := svc.Remove(entryB.ID); err == nil {
		t.Error("want ErrBookmarkNotFound on double remove, got nil")
	}
}

// TestBookmark_AddDedupSameLine 验证同一行重复添加不产生新记录，
// 且空 note/tag 的二次添加不覆盖已有备注。
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
		t.Fatalf("want same bookmark ID, got %d vs %d", first.ID, second.ID)
	}

	list, _ := svc.GetByFile(file)
	if len(list) != 1 {
		t.Fatalf("want 1 bookmark after dedup, got %d", len(list))
	}
	if list[0].Note != "impl here" {
		t.Errorf("note overwritten by empty add: want %q, got %q", "impl here", list[0].Note)
	}
}
