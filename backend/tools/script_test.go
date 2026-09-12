package tools

import (
	"strings"
	"sync"
	"testing"
	"time"
)

// shortenLuaTimeout 在测试内把脚本超时缩到 300ms，结束后恢复原值。
func shortenLuaTimeout(t *testing.T) {
	t.Helper()
	orig := luaTimeout
	luaTimeout = 300 * time.Millisecond
	t.Cleanup(func() { luaTimeout = orig })
}

// TestScript_Timeout 验证死循环脚本被超时拦截：返回失败结果且错误含"超时"，
// 不 panic、不双重 Close，超时窗口与配置一致。
func TestScript_Timeout(t *testing.T) {
	shortenLuaTimeout(t)
	svc := NewScriptService(t.TempDir())

	if err := svc.Save(ScriptInfo{ID: "infinite", Code: "while true do end", Enabled: true}); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	start := time.Now()
	result, err := svc.Execute("infinite", ScriptContext{})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Execute returned error (should be in result.Error): %v", err)
	}
	if result == nil || result.Success {
		t.Fatalf("expected Success=false, got %+v", result)
	}
	if !strings.Contains(result.Error, "超时") {
		t.Errorf("expected error to contain '超时', got: %q", result.Error)
	}
	if elapsed > 2*time.Second {
		t.Errorf("timeout took %v, expected ~300ms", elapsed)
	}
}

// TestScript_ConcurrentExecution 验证并发执行多个脚本（含死循环）安全：
// 正常脚本成功返回，死循环脚本超时失败，全程无 panic。
func TestScript_ConcurrentExecution(t *testing.T) {
	shortenLuaTimeout(t)
	svc := NewScriptService(t.TempDir())

	scripts := map[string]string{
		"fast1": `return 1`,
		"fast2": `return "hello"`,
		"slow1": `while true do end`,
		"slow2": `while true do end`,
	}
	ids := []string{"fast1", "fast2", "slow1", "slow2"}
	for id, code := range scripts {
		if err := svc.Save(ScriptInfo{ID: id, Code: code}); err != nil {
			t.Fatalf("Save %s: %v", id, err)
		}
	}

	var wg sync.WaitGroup
	results := make([]*ScriptResult, len(ids))
	errs := make([]error, len(ids))
	wg.Add(len(ids))
	for i, id := range ids {
		go func() {
			defer wg.Done()
			results[i], errs[i] = svc.Execute(id, ScriptContext{})
		}()
	}
	wg.Wait()

	for i, id := range ids {
		if errs[i] != nil || results[i] == nil {
			t.Fatalf("%s: err=%v result=%v", id, errs[i], results[i])
		}
		if strings.HasPrefix(id, "fast") && !results[i].Success {
			t.Errorf("%s expected Success=true, got error=%q", id, results[i].Error)
		}
		if strings.HasPrefix(id, "slow") && (results[i].Success || !strings.Contains(results[i].Error, "超时")) {
			t.Errorf("%s expected timeout failure, got %+v", id, results[i])
		}
	}
}
