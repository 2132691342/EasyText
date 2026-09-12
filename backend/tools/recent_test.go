package tools

import (
	"os"
	"path/filepath"
	"testing"

	"easy-text/backend/config"
)

// setupRecentTest 准备独立 DB + 配置源，limit>0 时覆盖最近文件上限。
func setupRecentTest(t *testing.T, limit int) *RecentService {
	t.Helper()
	tmpDir := t.TempDir()
	if err := config.InitDatabase(filepath.Join(tmpDir, "recent-test.db")); err != nil {
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
	return NewRecentService(config.DB, config.NewSource(config.Config))
}

// writeRecentFixture 写入临时文件并返回路径。
func writeRecentFixture(t *testing.T, dir, name string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

// TestRecent_Lifecycle 覆盖最近文件的核心业务规则：
// 同一文件重复打开去重（更新时间戳而非新增）、超过上限滚动淘汰旧记录、
// 文件与文件夹两个分类互不串扰。
func TestRecent_Lifecycle(t *testing.T) {
	svc := setupRecentTest(t, 3)
	defer closeDB(config.DB)

	// 去重：同一文件 Add 5 次只留 1 条
	dir := t.TempDir()
	dedupFile := writeRecentFixture(t, dir, "dedup.txt")
	for i := 0; i < 5; i++ {
		if err := svc.Add(dedupFile, false); err != nil {
			t.Fatalf("Add #%d: %v", i, err)
		}
	}
	list, err := svc.GetFiles()
	if err != nil || len(list) != 1 || list[0].Path != dedupFile {
		t.Fatalf("dedup: err=%v list=%+v", err, list)
	}

	// 上限：limit=3，第 4、5 个文件入列后最早的被淘汰
	var paths []string
	for _, name := range []string{"fA.txt", "fB.txt", "fC.txt", "fD.txt"} {
		p := writeRecentFixture(t, dir, name)
		paths = append(paths, p)
		if err := svc.Add(p, false); err != nil {
			t.Fatalf("Add %s: %v", p, err)
		}
	}
	list, err = svc.GetFiles()
	if err != nil {
		t.Fatalf("GetFiles: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("limit=3 expected, got %d entries", len(list))
	}
	last := paths[len(paths)-1]
	found := false
	for _, e := range list {
		if e.Path == last {
			found = true
		}
	}
	if !found {
		t.Errorf("latest entry missing; got %d entries", len(list))
	}

	// 文件/文件夹分类独立
	subdir := filepath.Join(dir, "subdir")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := svc.Add(subdir, true); err != nil {
		t.Fatalf("Add dir: %v", err)
	}
	files, err := svc.GetFiles()
	if err != nil {
		t.Fatal(err)
	}
	folders, err := svc.GetFolders()
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range files {
		if e.Path == subdir {
			t.Errorf("files leak folders: %q", e.Path)
		}
	}
	if len(folders) != 1 || folders[0].Path != subdir {
		t.Errorf("folders misclassified: %+v", folders)
	}
}
