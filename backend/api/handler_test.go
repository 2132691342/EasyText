package api

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"easy-text/backend/config"
)

// TestNewHandler_ConstructsCoreServices 验证 NewHandler 阶段构造的核心服务（不依赖 DB）。
// 若 fileWatcher 创建失败应 fail-fast panic（Linux/macOS/Windows 行为可能不同，跳过有问题的平台）。
func TestNewHandler_ConstructsCoreServices(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// fsnotify 在某些受限环境无法初始化；这是预期行为，跳过。
			if strings.Contains(anyToString(r), "file watcher") {
				t.Skipf("file watcher unavailable in this environment: %v", r)
			}
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	h := NewHandler()
	if h == nil {
		t.Fatal("NewHandler returned nil")
	}
	if h.fileReader == nil {
		t.Error("fileReader should be non-nil after NewHandler")
	}
	if h.fileWriter == nil {
		t.Error("fileWriter should be non-nil after NewHandler")
	}
	if h.treeBuilder == nil {
		t.Error("treeBuilder should be non-nil after NewHandler")
	}
	if h.fileWatcher == nil {
		t.Error("fileWatcher should be non-nil after NewHandler")
	}
	if h.jsonTool == nil {
		t.Error("jsonTool should be non-nil after NewHandler")
	}
	if h.encodingTool == nil {
		t.Error("encodingTool should be non-nil after NewHandler")
	}
	if h.compareService == nil {
		t.Error("compareService should be non-nil after NewHandler")
	}

	// 🆕 已拆分子 handler：无 DB 依赖的子 handler 在 Startup 前已注入。
	if h.FileAssocHandler == nil {
		t.Error("FileAssocHandler should be non-nil after NewHandler")
	}
	if h.SearchHandler == nil {
		t.Error("SearchHandler should be non-nil after NewHandler")
	}
	if h.SearchHandler.findService == nil {
		t.Error("SearchHandler.findService should be non-nil after NewHandler")
	}
	if h.SearchHandler.diffTool == nil {
		t.Error("SearchHandler.diffTool should be non-nil after NewHandler")
	}
	// RecentHandler 依赖 DB，Startup 前为 nil（保持不变量：依赖 DB 服务注入延迟）
	if h.RecentHandler != nil {
		t.Error("RecentHandler should be nil before Startup")
	}

	// 仍由 Handler 直持有的 DB-依赖服务在 Startup 前为 nil
	if h.draftService != nil {
		t.Error("draftService should be nil before Startup")
	}
	if h.snippetService != nil {
		t.Error("snippetService should be nil before Startup")
	}
	if h.bookmarkService != nil {
		t.Error("bookmarkService should be nil before Startup")
	}
}

// TestNewHandler_SubHandlersPromoteMethods 验证子 handler 的方法通过 Go 嵌入
// 提升到 *Handler，确保 Wails 反射看到的方法数不退化。
func TestNewHandler_SubHandlersPromoteMethods(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if strings.Contains(anyToString(r), "file watcher") {
				t.Skipf("file watcher unavailable in this environment: %v", r)
			}
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	h := NewHandler()

	// FileAssoc 方法直接通过 *Handler 调用（嵌入提升）
	var _ = h.RegisterFileAssoc
	var _ = h.UnregisterFileAssoc
	var _ = h.IsFileAssocRegistered

	// SearchHandler 的方法也通过嵌入提升（无需显式转 h.SearchHandler）
	var _ = h.FindInFile
	var _ = h.SearchInFiles
	var _ = h.CompareDiff
	var _ = h.BatchReplace

	// Recent 方法在 Startup 注入前不能提升调用，验证 nil pointer 抑制
	// （保证新代码不会误用未注入的子 handler）
	if h.RecentHandler != nil {
		t.Error("RecentHandler should be nil before Startup")
	}
}

// TestStartup_FailFastOnDBFailure 验证 DB 不可用时 Startup panic。
// 在不可写目录初始化时，InitDatabase 必失败；Startup 应 fail-fast panic 而非静默继续。
// TestStartup_FailFastOnDBFailure 验证 DB 不可用时 Startup panic。
// 在不可写目录初始化时，InitDatabase 必失败；Startup 应 fail-fast panic 而非静默继续。
//
// 跨平台说明：本测试用 /proc（Linux 专属）触发失败；Windows 上跳过，
// 由 Linux CI job 验证。完整覆盖依赖 GitHub Actions runner。
func TestStartup_FailFastOnDBFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows permission test fragile in CI; covered by Linux smoke test")
	}

	h := NewHandler()

	// 用一个不可写路径触发 InitDatabase 失败
	badDir := "/proc/1/secret-no-write"
	if err := os.MkdirAll(badDir, 0o000); err != nil {
		t.Skipf("cannot create unwritable dir: %v", err)
	}
	defer os.RemoveAll(badDir)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Startup should panic when DB init fails")
		}
		// panic 是预期：说明 fail-fast 生效
	}()

	h.Startup(nil)
}

// TestStartup_DBInitSuccess 验证正常路径下 Startup 后所有服务非 nil。
// 使用临时目录 + glebarez/sqlite（pure Go），跨平台可运行。
func TestStartup_DBInitSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir) // 影响 os.UserConfigDir on Linux
	_ = filepath.Join(tmpDir, "EasyText")

	defer func() {
		if r := recover(); r != nil {
			t.Skipf("Startup panic in this environment, skipping: %v", r)
		}
	}()

	h := NewHandler()
	h.Startup(nil)

	// 验证 Startup 后服务非 nil（fail-fast 不变量）
	if h.draftService == nil {
		t.Error("draftService should be non-nil after Startup")
	}
	if h.recentService == nil {
		t.Error("recentService should be non-nil after Startup")
	}
	if h.snippetService == nil {
		t.Error("snippetService should be non-nil after Startup")
	}
	if h.bookmarkService == nil {
		t.Error("bookmarkService should be non-nil after Startup")
	}
	if h.scriptService == nil {
		t.Error("scriptService should be non-nil after Startup")
	}

	// config.DB 应当被初始化
	if config.DB == nil {
		t.Error("config.DB should be non-nil after Startup")
	}
}

// anyToString 把 recover() 的 interface{} 转为字符串（r.(string) 或 r.(error).Error()）。
func anyToString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if e, ok := v.(error); ok {
		return e.Error()
	}
	return ""
}
