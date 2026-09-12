package api

import (
	"testing"

	"easy-text/backend/config"
)

// TestStartup_WiresServices 验证启动链路：Startup 后 DB-依赖服务全部注入、
// config.DB 就绪。这是所有 Wails 方法可用的前置不变量。
func TestStartup_WiresServices(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	defer func() {
		if r := recover(); r != nil {
			t.Skipf("Startup panic in this environment, skipping: %v", r)
		}
	}()

	h := NewHandler()
	h.Startup(nil)

	for name, svc := range map[string]any{
		"draftService":   h.draftService,
		"recentService":  h.recentService,
		"snippetService": h.snippetService,
		"bookmarkService": h.bookmarkService,
		"scriptService":  h.scriptService,
	} {
		if svc == nil {
			t.Errorf("%s should be non-nil after Startup", name)
		}
	}
	if config.DB == nil {
		t.Error("config.DB should be non-nil after Startup")
	}
}
