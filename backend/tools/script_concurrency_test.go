package tools

import (
	"strings"
	"sync"
	"testing"
)

// TestExecuteLua_ConcurrentNoPanic 验证并发执行多个 Lua 脚本不触发 panic。
// F-1 修复前若 timeout 分支双重 Close 会出现 panic；修复后应安全。
func TestExecuteLua_ConcurrentNoPanic(t *testing.T) {
	svc := NewScriptService(t.TempDir())

	// 准备 5 个脚本：3 个正常 + 2 个死循环
	scripts := map[string]string{
		"fast1": `return 1`,
		"fast2": `return 2`,
		"fast3": `return 3`,
		"slow1": `while true do end`,
		"slow2": `while true do end`,
	}
	for id, code := range scripts {
		if err := svc.Save(ScriptInfo{ID: id, Code: code}); err != nil {
			t.Fatalf("Save %s: %v", id, err)
		}
	}

	var wg sync.WaitGroup
	results := make([]*ScriptResult, len(scripts))
	errs := make([]error, len(scripts))

	wg.Add(len(scripts))
	for i, id := range []string{"fast1", "fast2", "fast3", "slow1", "slow2"} {
		i, id := i, id
		go func() {
			defer wg.Done()
			results[i], errs[i] = svc.Execute(id, ScriptContext{})
		}()
	}
	wg.Wait()

	for i, id := range []string{"fast1", "fast2", "fast3", "slow1", "slow2"} {
		if errs[i] != nil {
			t.Errorf("%s returned error: %v", id, errs[i])
		}
		if results[i] == nil {
			t.Errorf("%s nil result", id)
			continue
		}
		if id == "fast1" || id == "fast2" || id == "fast3" {
			if !results[i].Success {
				t.Errorf("%s expected Success=true, got error=%q", id, results[i].Error)
			}
		} else {
			if results[i].Success {
				t.Errorf("%s expected Success=false", id)
			}
			if !strings.Contains(results[i].Error, "超时") {
				t.Errorf("%s expected error to contain '超时', got %q", id, results[i].Error)
			}
		}
	}
}
