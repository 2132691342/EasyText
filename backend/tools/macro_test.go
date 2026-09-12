package tools

import (
	"path/filepath"
	"testing"
)

// TestMacro_RecordAndStop 验证录制生命周期：Start → RecordStep ×N → Stop 产出宏。
func TestMacro_RecordAndStop(t *testing.T) {
	svc := NewMacroService()
	// 隔离持久化：测试写到 TempDir，避免污染用户配置目录
	svc.storagePath = filepath.Join(t.TempDir(), "macros.json")

	if svc.IsRecording() {
		t.Fatal("recording should start off")
	}
	svc.StartRecording()
	if !svc.IsRecording() {
		t.Fatal("want recording=true after StartRecording")
	}

	svc.RecordStep(MacroStep{Type: "insert", Text: "abc", Timestamp: 1})
	svc.RecordStep(MacroStep{Type: "delete", Text: "a", Timestamp: 2})

	macro := svc.StopRecording()
	if macro == nil {
		t.Fatal("want macro after StopRecording, got nil")
	}
	if len(macro.Steps) != 2 {
		t.Errorf("want 2 steps, got %d", len(macro.Steps))
	}
	if svc.IsRecording() {
		t.Error("recording should stop after StopRecording")
	}
}

// TestMacro_StopWithoutRecording 验证未录制时 Stop 返回 nil。
func TestMacro_StopWithoutRecording(t *testing.T) {
	svc := NewMacroService()
	svc.storagePath = filepath.Join(t.TempDir(), "macros.json")

	if m := svc.StopRecording(); m != nil {
		t.Errorf("want nil macro without recording, got %+v", m)
	}
}

// TestMacro_GetDeleteRename 验证宏查询 / 删除 / 重命名。
func TestMacro_GetDeleteRename(t *testing.T) {
	svc := NewMacroService()
	svc.storagePath = filepath.Join(t.TempDir(), "macros.json")

	svc.StartRecording()
	svc.RecordStep(MacroStep{Type: "insert", Text: "x", Timestamp: 1})
	saved := svc.SaveCurrentMacro("first")
	if saved == nil {
		t.Fatal("want saved macro, got nil")
	}

	all := svc.GetMacros()
	if len(all) != 1 {
		t.Fatalf("want 1 macro, got %d", len(all))
	}

	if got := svc.GetMacro(saved.ID); got == nil || got.Name != "first" {
		t.Errorf("GetMacro mismatch: %+v", got)
	}
	if !svc.RenameMacro(saved.ID, "renamed") {
		t.Error("RenameMacro should succeed for existing id")
	}
	if got := svc.GetMacro(saved.ID); got.Name != "renamed" {
		t.Errorf("want name=renamed, got %q", got.Name)
	}
	if !svc.DeleteMacro(saved.ID) {
		t.Error("DeleteMacro should succeed for existing id")
	}
	if svc.DeleteMacro(saved.ID) {
		t.Error("double delete should report false")
	}
}

// TestMacro_SaveCurrentRequiresSteps 验证空录制不能保存（防产生空宏）。
func TestMacro_SaveCurrentRequiresSteps(t *testing.T) {
	svc := NewMacroService()
	svc.storagePath = filepath.Join(t.TempDir(), "macros.json")

	svc.StartRecording()
	if m := svc.SaveCurrentMacro("empty"); m != nil {
		t.Errorf("want nil for empty recording, got %+v", m)
	}
}

// TestMacro_PersistenceRoundTrip 验证宏文件保存后重建实例可恢复（防回归：宏丢失）。
func TestMacro_PersistenceRoundTrip(t *testing.T) {
	dir := t.TempDir()
	storage := filepath.Join(dir, "macros.json")

	svcA := NewMacroService()
	svcA.storagePath = storage
	svcA.StartRecording()
	svcA.RecordStep(MacroStep{Type: "insert", Text: "persist", Timestamp: 1})
	if m := svcA.SaveCurrentMacro("keep-me"); m == nil {
		t.Fatal("SaveCurrentMacro failed")
	}

	// 新实例从磁盘 load：构造函数已按 UserConfigDir load 过一次，
	// 覆盖 storagePath 后需再显式 load 一次（同包测试可直接调私有方法）
	svcB := NewMacroService()
	svcB.storagePath = storage
	svcB.load()
	all := svcB.GetMacros()
	if len(all) != 1 || all[0].Name != "keep-me" {
		t.Fatalf("persistence round-trip failed: %+v", all)
	}
}
