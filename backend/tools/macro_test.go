package tools

import (
	"path/filepath"
	"testing"
)

// TestMacro_Lifecycle 覆盖宏的完整生命周期：录制 → 步骤采集 → 停止产出宏 →
// 保存（空录制必须拒绝）→ 查询/重命名/删除，以及删除不存在 ID 的幂等语义。
func TestMacro_Lifecycle(t *testing.T) {
	svc := NewMacroService()
	svc.storagePath = filepath.Join(t.TempDir(), "macros.json")

	if svc.IsRecording() {
		t.Fatal("recording should start off")
	}
	svc.StartRecording()
	svc.RecordStep(MacroStep{Type: "insert", Text: "abc", Timestamp: 1})
	svc.RecordStep(MacroStep{Type: "delete", Text: "a", Timestamp: 2})

	// 保存即结束录制（内部清空缓冲并置 isRecording=false）
	saved := svc.SaveCurrentMacro("first")
	if saved == nil || len(saved.Steps) != 2 {
		t.Fatalf("SaveCurrentMacro should yield 2 steps, got %+v", saved)
	}
	if svc.IsRecording() {
		t.Error("recording should stop after SaveCurrentMacro")
	}
	if got := svc.GetMacro(saved.ID); got == nil || got.Name != "first" {
		t.Errorf("GetMacro mismatch: %+v", got)
	}
	if !svc.RenameMacro(saved.ID, "renamed") || svc.GetMacro(saved.ID).Name != "renamed" {
		t.Error("RenameMacro failed")
	}
	if !svc.DeleteMacro(saved.ID) || svc.DeleteMacro(saved.ID) {
		t.Error("delete should succeed once then report false")
	}

	// StopRecording 同样保存宏（自动命名）
	svc.StartRecording()
	svc.RecordStep(MacroStep{Type: "insert", Text: "x", Timestamp: 3})
	if m := svc.StopRecording(); m == nil || len(m.Steps) != 1 {
		t.Fatalf("StopRecording should yield 1 step, got %+v", m)
	}
	if all := svc.GetMacros(); len(all) != 1 {
		t.Errorf("stopped macro should be saved, got %d", len(all))
	}

	// 空录制不允许保存为宏
	empty := NewMacroService()
	empty.storagePath = filepath.Join(t.TempDir(), "macros.json")
	empty.StartRecording()
	if m := empty.SaveCurrentMacro("empty"); m != nil {
		t.Errorf("empty recording must not save, got %+v", m)
	}
}

// TestMacro_PersistenceRoundTrip 验证宏文件落盘后新实例可恢复（防宏丢失回归）。
func TestMacro_PersistenceRoundTrip(t *testing.T) {
	storage := filepath.Join(t.TempDir(), "macros.json")

	svcA := NewMacroService()
	svcA.storagePath = storage
	svcA.StartRecording()
	svcA.RecordStep(MacroStep{Type: "insert", Text: "persist", Timestamp: 1})
	if m := svcA.SaveCurrentMacro("keep-me"); m == nil {
		t.Fatal("SaveCurrentMacro failed")
	}

	svcB := NewMacroService()
	svcB.storagePath = storage
	svcB.load()
	all := svcB.GetMacros()
	if len(all) != 1 || all[0].Name != "keep-me" {
		t.Fatalf("persistence round-trip failed: %+v", all)
	}
}
