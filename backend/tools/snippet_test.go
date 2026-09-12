package tools

import (
	"encoding/json"
	"strings"
	"testing"

	"easy-text/backend/config"
)

// TestSnippet_CRUDAndLanguageFilter 覆盖片段增改删，以及语言过滤语义：
// 匹配语言与空语言（通用片段）都返回，不匹配的过滤掉。
func TestSnippet_CRUDAndLanguageFilter(t *testing.T) {
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

	for lang, want := range map[string]int{"go": 2, "javascript": 2, "": 3} {
		list, err := svc.GetAll(lang)
		if err != nil {
			t.Fatalf("GetAll(%q): %v", lang, err)
		}
		if len(list) != want {
			t.Errorf("GetAll(%q): want %d, got %d", lang, want, len(list))
		}
	}

	list, _ := svc.GetAll("")
	list[0].Body = "updated body"
	if err := svc.Update(&list[0]); err != nil {
		t.Fatalf("Update: %v", err)
	}
	after, _ := svc.GetAll("")
	updated := false
	for _, sn := range after {
		if sn.Body == "updated body" {
			updated = true
		}
	}
	if !updated {
		t.Error("Update did not persist body")
	}

	for _, sn := range after {
		if err := svc.Delete(sn.ID); err != nil {
			t.Fatalf("Delete %d: %v", sn.ID, err)
		}
	}
	if final, _ := svc.GetAll(""); len(final) != 0 {
		t.Errorf("want 0 after delete all, got %d", len(final))
	}
}

// TestSnippet_ImportExportRoundTrip 覆盖导入/导出长链路：VS Code 格式
// （body 支持 []string 拼接、prefix 缺省回退到 name）、EasyText 数组格式、
// 非法 JSON 报错，以及导出 JSON 可被重新解析且内容一致。
func TestSnippet_ImportExportRoundTrip(t *testing.T) {
	cleanup := setupDBTest(t)
	defer cleanup()

	svc := NewSnippetService(config.DB)

	vsJSON := `{
		"print log": {"prefix": "plog", "body": ["line1", "line2"], "description": "two-line body"},
		"no-prefix": {"body": "uses name as prefix"}
	}`
	count, err := svc.ImportFromJSON(vsJSON)
	if err != nil || count != 2 {
		t.Fatalf("ImportFromJSON(vscode): count=%d err=%v", count, err)
	}
	list, _ := svc.GetAll("")
	byName := map[string]SnippetEntry{}
	for _, sn := range list {
		byName[sn.Name] = sn
	}
	if got := byName["print log"]; got.Body != "line1\nline2" {
		t.Errorf("[]string body join failed: %q", got.Body)
	}
	if got := byName["no-prefix"]; got.Prefix != "no-prefix" {
		t.Errorf("prefix fallback to name failed: %q", got.Prefix)
	}

	if _, err := svc.ImportFromJSON(`[{"name":"n1","prefix":"p1","body":"b1","language":"go"}]`); err != nil {
		t.Fatalf("ImportFromJSON(easytext): %v", err)
	}
	if _, err := svc.ImportFromJSON(`{invalid json`); err == nil {
		t.Error("want error for invalid JSON, got nil")
	}

	exported, err := svc.ExportToJSON()
	if err != nil {
		t.Fatalf("ExportToJSON: %v", err)
	}
	var entries []SnippetEntry
	if err := json.Unmarshal([]byte(exported), &entries); err != nil {
		t.Fatalf("reparse exported JSON: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("want 3 entries after round-trip, got %d (%s)", len(entries), strings.TrimSpace(exported))
	}
}
