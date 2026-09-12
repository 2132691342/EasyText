package tools

import (
	"encoding/json"
	"strings"
	"testing"

	"easy-text/backend/config"
)

// TestSnippet_CreateUpdateDelete CRUD 全链路。
func TestSnippet_CreateUpdateDelete(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewSnippetService(config.DB)

	id, err := svc.Create(&SnippetEntry{
		Name: "logger", Prefix: "logx", Body: "console.log($0)",
		Description: "log helper", Language: "javascript",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("want positive id, got %d", id)
	}

	list, err := svc.GetAll("")
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(list) != 1 || list[0].Prefix != "logx" {
		t.Fatalf("after create: want [logx], got %+v", list)
	}

	list[0].Body = "log.debug($0)"
	if err := svc.Update(&list[0]); err != nil {
		t.Fatalf("Update: %v", err)
	}
	after, _ := svc.GetAll("")
	if after[0].Body != "log.debug($0)" {
		t.Errorf("body not updated: %q", after[0].Body)
	}

	if err := svc.Delete(id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	final, _ := svc.GetAll("")
	if len(final) != 0 {
		t.Errorf("want 0 after delete, got %d", len(final))
	}
}

// TestSnippet_GetAllLanguageFilter 验证语言过滤：
// 匹配语言 + 空语言（通用片段）都会返回；不匹配的过滤掉。
func TestSnippet_GetAllLanguageFilter(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewSnippetService(config.DB)
	fixtures := []SnippetEntry{
		{Name: "go-fn", Prefix: "gfn", Body: "func x() {}", Language: "go"},
		{Name: "js-fn", Prefix: "jfn", Body: "function x() {}", Language: "javascript"},
		{Name: "common", Prefix: "hdr", Body: "// header", Language: ""},
	}
	for i := range fixtures {
		if _, err := svc.Create(&fixtures[i]); err != nil {
			t.Fatalf("Create %s: %v", fixtures[i].Name, err)
		}
	}

	goSnippets, err := svc.GetAll("go")
	if err != nil {
		t.Fatalf("GetAll(go): %v", err)
	}
	if len(goSnippets) != 2 { // go-fn + 通用 common
		t.Errorf("want 2 for lang=go, got %d", len(goSnippets))
	}
	jsSnippets, _ := svc.GetAll("javascript")
	if len(jsSnippets) != 2 { // js-fn + 通用 common
		t.Errorf("want 2 for lang=javascript, got %d", len(jsSnippets))
	}
	all, _ := svc.GetAll("")
	if len(all) != 3 {
		t.Errorf("want 3 for lang='', got %d", len(all))
	}
}

// TestSnippet_ImportVSCodeFormat 验证 VS Code 格式导入：
// map[name]{prefix, body(string 或 []string)}，prefix 缺省用 name。
func TestSnippet_ImportVSCodeFormat(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewSnippetService(config.DB)
	vsJSON := `{
		"print log": {
			"prefix": "plog",
			"body": ["line1", "line2"],
			"description": "two-line body"
		},
		"simple": {
			"prefix": "s",
			"body": "single string body"
		},
		"no-prefix": {
			"body": "uses name as prefix"
		}
	}`
	count, err := svc.ImportFromJSON(vsJSON)
	if err != nil {
		t.Fatalf("ImportFromJSON: %v", err)
	}
	if count != 3 {
		t.Fatalf("want 3 imported, got %d", count)
	}

	list, _ := svc.GetAll("")
	byName := map[string]SnippetEntry{}
	for _, sn := range list {
		byName[sn.Name] = sn
	}
	if got := byName["print log"]; got.Body != "line1\nline2" {
		t.Errorf("[]string body join: want %q, got %q", "line1\nline2", got.Body)
	}
	if got := byName["no-prefix"]; got.Prefix != "no-prefix" {
		t.Errorf("prefix fallback to name: got %q", got.Prefix)
	}
}

// TestSnippet_ImportEasyTextFormat 验证 EasyText 数组格式导入与无效格式报错。
func TestSnippet_ImportEasyTextFormat(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewSnippetService(config.DB)
	etJSON := `[{"name":"n1","prefix":"p1","body":"b1","language":"go"}]`
	count, err := svc.ImportFromJSON(etJSON)
	if err != nil {
		t.Fatalf("Import easytext format: %v", err)
	}
	if count != 1 {
		t.Fatalf("want 1 imported, got %d", count)
	}

	if _, err := svc.ImportFromJSON(`{invalid json`); err == nil {
		t.Error("want error for invalid JSON, got nil")
	}
}

// TestSnippet_ExportRoundTrip 验证导出为合法 JSON 且导入-导出内容一致。
func TestSnippet_ExportRoundTrip(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewSnippetService(config.DB)
	if _, err := svc.ImportFromJSON(`[{"name":"x","prefix":"xp","body":"xb","language":"go"}]`); err != nil {
		t.Fatalf("Import: %v", err)
	}

	exported, err := svc.ExportToJSON()
	if err != nil {
		t.Fatalf("ExportToJSON: %v", err)
	}
	if !strings.Contains(exported, `"prefix": "xp"`) {
		t.Errorf("exported JSON missing prefix xp: %s", exported)
	}

	// 导出的 JSON 可被重新解析为 entries 数组
	var entries []SnippetEntry
	if err := json.Unmarshal([]byte(exported), &entries); err != nil {
		t.Fatalf("reparse exported JSON: %v", err)
	}
	if len(entries) != 1 || entries[0].Prefix != "xp" {
		t.Errorf("roundtrip mismatch: %+v", entries)
	}
}
