package tools

import (
	"strings"
	"testing"
	"time"
)

// TestExecuteLua_TimeoutNoPanic 复现 F-1 修复前的 panic 风险：
// 超时分支 Close + 外层 defer Close → 双重 Close panic。
//
// 修复后：L.Close 由执行 goroutine 独占负责，外层 select 超时分支不再调 Close。
// 即使脚本是死循环，超时后应返回 ScriptResult{Success:false, Error:含"超时"}，
// 不应 panic，也不应让后续调用受影响。
func TestExecuteLua_TimeoutNoPanic(t *testing.T) {
	svc := NewScriptService(t.TempDir())

	// 死循环脚本
	infiniteLoop := `
		while true do
			-- 永远不退出
		end
	`

	// 通过 Execute() 路径调用，先保存脚本再执行。
	// 直接调用 executeLua 需要 ScriptInfo，不便；用 Execute 通过 Lua 文件。
	if err := svc.Save(ScriptInfo{
		ID:      "infinite",
		Code:    infiniteLoop,
		Enabled: true,
	}); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	start := time.Now()
	result, err := svc.Execute("infinite", ScriptContext{})
	elapsed := time.Since(start)

	// 不应 panic；err 应为 nil（错误信息已在 result.Error 中）
	if err != nil {
		t.Fatalf("Execute returned error (should be in result.Error): %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if result.Success {
		t.Errorf("expected Success=false for infinite loop, got true")
	}
	if !strings.Contains(result.Error, "超时") {
		t.Errorf("expected error to contain '超时', got: %q", result.Error)
	}

	// 验证超时窗口合理（luaTimeout = 5s，允许 ±1s 抖动）
	if elapsed < 4*time.Second || elapsed > 7*time.Second {
		t.Errorf("timeout took %v, expected ~5s", elapsed)
	}
}

// TestExecuteLua_NormalCompletion 验证正常脚本能完成。
func TestExecuteLua_NormalCompletion(t *testing.T) {
	svc := NewScriptService(t.TempDir())
	if err := svc.Save(ScriptInfo{
		ID:   "ok",
		Code: `return "hello"`,
	}); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	result, err := svc.Execute("ok", ScriptContext{Content: "world"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !result.Success {
		t.Errorf("expected Success=true, got: %v (error=%q)", result.Success, result.Error)
	}
}
