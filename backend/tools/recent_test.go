package tools

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"easy-text/backend/config"
	"gorm.io/gorm"
)

// setupRecentTest 准备一次性的 DB（每个 subtest 用各自的 t.TempDir 文件）。
// 由于 config.DB 是 package-global，gorm 持有 sqlite 连接句柄，
// 测试结束时需要先关闭连接，再让 t.TempDir 收回目录。
func setupRecentTest(t *testing.T, limit int) (*RecentService, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "recent-test.db")

	if err := config.InitDatabase(dbPath); err != nil {
		t.Fatalf("InitDatabase: %v", err)
	}
	if err := config.InitConfig(filepath.Join(tmpDir, "config.json")); err != nil {
		t.Fatalf("InitConfig: %v", err)
	}
	if limit > 0 {
		cfg := config.Config.Get()
		cfg.UI.RecentFilesLimit = limit
		if err := config.Config.Update(cfg); err != nil {
			t.Fatalf("Config.Update: %v", err)
		}
	}

	svc := NewRecentService(config.DB, config.NewSource(config.Config))
	cleanup := func() {
		closeDB(config.DB)
	}
	return svc, cleanup
}

// closeDB 关闭 gorm DB 持有的 sql.DB 连接，确保临时目录在 defer 时可删除。
// 在 Windows 上，若文件句柄未释放，os.RemoveAll(TempDir) 会失败。
func closeDB(db *gorm.DB) {
	if db == nil {
		return
	}
	if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
		_ = sqlDB.Close()
	}
}

// TestRecent_AddThenGet 验证基本写入与读取：先 Add 一条，再 GetFiles 取出。
func TestRecent_AddThenGet(t *testing.T) {
	svc, cleanup := setupRecentTest(t, 10)
	defer cleanup()

	tmpFile := filepath.Join(t.TempDir(), "alpha.txt")
	if err := os.WriteFile(tmpFile, []byte("alpha"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := svc.Add(tmpFile, false); err != nil {
		t.Fatalf("Add: %v", err)
	}

	list, err := svc.GetFiles()
	if err != nil {
		t.Fatalf("GetFiles: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 entry, got %d", len(list))
	}
	if list[0].Path != tmpFile {
		t.Errorf("path mismatch: want %q, got %q", tmpFile, list[0].Path)
	}
}

// TestRecent_DedupOnReopen 验证同一文件重复 Add 会更新时间戳而非新增记录，
// 修复「打开同一文件多次产生重复最近记录」的潜在回归。
func TestRecent_DedupOnReopen(t *testing.T) {
	svc, cleanup := setupRecentTest(t, 10)
	defer cleanup()

	tmpFile := filepath.Join(t.TempDir(), "dedup.txt")
	if err := os.WriteFile(tmpFile, []byte("x"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	for i := 0; i < 5; i++ {
		if err := svc.Add(tmpFile, false); err != nil {
			t.Fatalf("Add #%d: %v", i, err)
		}
	}

	list, err := svc.GetFiles()
	if err != nil {
		t.Fatalf("GetFiles: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("repeated Add must dedup to 1 entry, got %d", len(list))
	}
}

// TestRecent_LimitEnforced 验证超过 limit 的旧记录被自动清理。
// 修复「未滚动保留导致最近文件无限增长」的 Bug：默认 10 条，超出后入新删旧。
func TestRecent_LimitEnforced(t *testing.T) {
	svc, cleanup := setupRecentTest(t, 3)
	defer cleanup()

	tmpDir := t.TempDir()
	// 写入 5 个不同的文件，超过 limit=3
	paths := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		p := filepath.Join(tmpDir, "f"+string(rune('A'+i))+".txt")
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
		paths = append(paths, p)
		if err := svc.Add(p, false); err != nil {
			t.Fatalf("Add %s: %v", p, err)
		}
	}

	list, err := svc.GetFiles()
	if err != nil {
		t.Fatalf("GetFiles: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("limit=3 expected, got %d entries: %v", len(list), pathsOf(list))
	}
	// 最新的两条应在（最后一次 Add 会 bump 时间戳）
	last := paths[len(paths)-1]
	foundLast := false
	for _, e := range list {
		if e.Path == last {
			foundLast = true
		}
	}
	if !foundLast {
		t.Errorf("latest entry missing; got %v", pathsOf(list))
	}
}

// TestRecent_FileAndFolderSeparate 验证文件 / 文件夹分类独立，互不干扰。
func TestRecent_FileAndFolderSeparate(t *testing.T) {
	svc, cleanup := setupRecentTest(t, 10)
	defer cleanup()

	tmpDir := t.TempDir()
	fileA := filepath.Join(tmpDir, "a.txt")
	fileB := filepath.Join(tmpDir, "b.txt")
	if err := os.WriteFile(fileA, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fileB, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := svc.Add(fileA, false); err != nil {
		t.Fatal(err)
	}
	if err := svc.Add(fileB, false); err != nil {
		t.Fatal(err)
	}
	if err := svc.Add(dir, true); err != nil {
		t.Fatal(err)
	}

	files, err := svc.GetFiles()
	if err != nil {
		t.Fatal(err)
	}
	folders, err := svc.GetFolders()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0].Path == dir || files[1].Path == dir {
		t.Errorf("files leak folders: %v", pathsOf(files))
	}
	if len(folders) != 1 || folders[0].Path != dir {
		t.Errorf("folders misclassified: %v", pathsOf(folders))
	}
}

// TestRecent_DefaultLimitIs10 防御：默认上限是 10，与 UI 默认 recentFilesLimit 对齐。
// 若有人不慎改回 20，UI/后端会再次错位——这条测试能立刻发现。
func TestRecent_DefaultLimitIs10(t *testing.T) {
	if defaultRecentLimit != 10 {
		t.Errorf("defaultRecentLimit changed: want 10, got %d", defaultRecentLimit)
	}
}

// fakeConfigSource 是 Step 5（ch11 DIP）证明用：业务服务只通过 Source 接口
// 拿配置，无需 global config.Config 也能跑测试。
type fakeConfigSource struct {
	limit int
	err   error
}

func (f *fakeConfigSource) Get() config.AppConfig {
	if f.err != nil {
		// 故意让 Get() 返回空配置而不 panic——业务代码应能容忍 Source 异常。
		return config.AppConfig{}
	}
	return config.AppConfig{
		UI: config.UIConfig{RecentFilesLimit: f.limit},
	}
}

// TestRecent_NewRecentService_NilCfgSource_FallsBackToNoop 防御：忘记注入 Source
// 时业务服务不能 panic。这条测试确保 graceful degradation。
func TestRecent_NewRecentService_NilCfgSource_FallsBackToNoop(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "nil-cfg.db")
	if err := config.InitDatabase(dbPath); err != nil {
		t.Fatal(err)
	}
	defer closeDB(config.DB)

	// 不传 cfgSource → 走 noop 默认值（RecentFilesLimit=0 → defaultRecentLimit）
	svc := NewRecentService(config.DB, nil)
	if svc == nil {
		t.Fatal("svc should not be nil")
	}
	// 调用一次，确认能跑通
	tmpFile := filepath.Join(t.TempDir(), "x.txt")
	_ = os.WriteFile(tmpFile, []byte("x"), 0o644)
	if err := svc.Add(tmpFile, false); err != nil {
		t.Fatalf("Add with nil cfgSource: %v", err)
	}
}

// TestRecent_DI_FakeSource 验证 DI 路径：通过 fakeConfigSource 注入 limit=2，
// 业务行为走 fake 路径，不再依赖全局 config.Config。这是 Step 5 的核心 ROI。
func TestRecent_DI_FakeSource(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "di.db")
	if err := config.InitDatabase(dbPath); err != nil {
		t.Fatal(err)
	}
	defer closeDB(config.DB)

	fake := &fakeConfigSource{limit: 2}
	svc := NewRecentService(config.DB, fake)

	// 写 5 个文件，超过 limit=2
	tmp := t.TempDir()
	for i := 0; i < 5; i++ {
		p := filepath.Join(tmp, intToStr(i)+".txt")
		_ = os.WriteFile(p, []byte("x"), 0o644)
		_ = svc.Add(p, false)
	}

	res, err := svc.GetFiles()
	if err != nil {
		t.Fatalf("GetFiles: %v", err)
	}
	if got := len(res); got != 2 {
		t.Errorf("DI-based limit: want 2, got %d", got)
	}
}

// TestRecent_DI_SourceErrorTolerated 防御：Source.Get 返回错误时业务层不应 panic。
// （实际上 fakeConfigSource 返回空 config.AppConfig{}；这里仅做类型层校验。）
func TestRecent_DI_SourceErrorTolerated(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "src-err.db")
	if err := config.InitDatabase(dbPath); err != nil {
		t.Fatal(err)
	}
	defer closeDB(config.DB)

	fake := &fakeConfigSource{err: errors.New("source broken")}
	svc := NewRecentService(config.DB, fake)

	tmpFile := filepath.Join(t.TempDir(), "y.txt")
	_ = os.WriteFile(tmpFile, []byte("y"), 0o644)
	// 不应 panic；调用走空配置路径（Limit=0 → fallback 10）
	_ = svc.Add(tmpFile, false)
}

// pathsOf 将 results 数组的 path 列抽出来，便于调试信息。
func pathsOf(list []RecentEntryResult) []string {
	out := make([]string, 0, len(list))
	for _, e := range list {
		out = append(out, e.Path)
	}
	return out
}
