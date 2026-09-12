package tools

import (
	"strings"
	"testing"
)

// TestJSON_FormatAndMinify 验证格式化（含缩进）与压缩往返。
func TestJSON_FormatAndMinify(t *testing.T) {
	jt := NewJSONTool()

	r := jt.Format(`{"a":1,"b":[2,3]}`, 2)
	if !r.Success {
		t.Fatalf("Format failed: %+v", r.Error)
	}
	if !strings.Contains(r.Content, "\n") || !strings.Contains(r.Content, `"a": 1`) {
		t.Errorf("format output unexpected: %q", r.Content)
	}

	m := jt.Minify(r.Content)
	if !m.Success {
		t.Fatalf("Minify failed: %+v", m.Error)
	}
	if strings.Contains(m.Content, "\n") {
		t.Errorf("minified should be single line: %q", m.Content)
	}
}

// TestJSON_ValidateInvalid 验证非法 JSON 返回 Success=false 且错误带行号信息。
func TestJSON_ValidateInvalid(t *testing.T) {
	jt := NewJSONTool()

	bad := "{\n  \"a\": 1,\n  \"b\": \n}"
	r := jt.Validate(bad)
	if r.Success {
		t.Fatalf("want invalid JSON to fail, got %+v", r)
	}
	if r.Error == nil {
		t.Fatal("want error details, got nil")
	}
	if r.Error.Line <= 0 {
		t.Errorf("want line number in error, got %d", r.Error.Line)
	}
}

// TestJSON_QueryPath 验证 JSONPath 查询命中与未命中。
func TestJSON_QueryPath(t *testing.T) {
	jt := NewJSONTool()
	content := `{"store":{"book":[{"title":"a"},{"title":"b"}]}}`

	results, err := jt.QueryPath(content, "$.store.book[*].title")
	if err != nil {
		t.Fatalf("QueryPath: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 hits, got %d", len(results))
	}

	results, err = jt.QueryPath(content, "$.nothing.here")
	if err != nil {
		t.Fatalf("QueryPath(miss): %v", err)
	}
	if len(results) != 0 {
		t.Errorf("want 0 hits, got %d", len(results))
	}
}

// TestJSON_GenerateStruct 验证 Go / TypeScript 结构体生成包含字段与类型。
func TestJSON_GenerateStruct(t *testing.T) {
	jt := NewJSONTool()
	content := `{"id":1,"name":"x","active":true}`

	goCode, err := jt.GenerateStruct(content, "go", "User")
	if err != nil {
		t.Fatalf("GenerateStruct(go): %v", err)
	}
	if !strings.Contains(goCode, "type User struct") || !strings.Contains(goCode, "Id") {
		t.Errorf("go struct unexpected: %s", goCode)
	}

	tsCode, err := jt.GenerateStruct(content, "typescript", "User")
	if err != nil {
		t.Fatalf("GenerateStruct(ts): %v", err)
	}
	if !strings.Contains(tsCode, "interface User") || !strings.Contains(tsCode, "name: string") {
		t.Errorf("ts interface unexpected: %s", tsCode)
	}
}

// TestJSON_ExtractKeysAndFlatten 验证键提取去重与 Flatten 嵌套展开。
func TestJSON_ExtractKeysAndFlatten(t *testing.T) {
	jt := NewJSONTool()
	content := `{"a":1,"b":{"c":2,"d":{"e":3}},"a2":4}`

	keys, err := jt.ExtractKeys(content)
	if err != nil {
		t.Fatalf("ExtractKeys: %v", err)
	}
	for _, want := range []string{"a", "b", "c", "d", "e"} {
		if !containsStr(keys, want) {
			t.Errorf("key %q missing from %v", want, keys)
		}
	}

	flat, err := jt.Flatten(content, ".")
	if err != nil {
		t.Fatalf("Flatten: %v", err)
	}
	if !strings.Contains(flat, `"b.d.e"`) && !strings.Contains(flat, "b.d.e") {
		t.Errorf("flatten missing nested path: %s", flat)
	}
}

// TestJSON_StructuredDiff 验证结构化 Diff 的 added/removed/modified 分类。
func TestJSON_StructuredDiff(t *testing.T) {
	jt := NewJSONTool()
	left := `{"a":1,"b":"keep","c":true}`
	right := `{"a":2,"b":"keep","d":"new"}`

	res, err := jt.StructuredDiff(left, right)
	if err != nil {
		t.Fatalf("StructuredDiff: %v", err)
	}
	// Path 形如 "$.a"（实现以 "$" 为根，"$.key" 为子路径）
	kinds := map[string]string{}
	for _, e := range res.Entries {
		kinds[e.Path] = e.Type
	}
	if kinds["$.a"] != "modified" {
		t.Errorf("a should be modified, got %q", kinds["$.a"])
	}
	if kinds["$.c"] != "removed" {
		t.Errorf("c should be removed, got %q", kinds["$.c"])
	}
	if kinds["$.d"] != "added" {
		t.Errorf("d should be added, got %q", kinds["$.d"])
	}
	if kinds["$.b"] != "unchanged" && kinds["$.b"] != "" {
		t.Errorf("b should be unchanged, got %q", kinds["$.b"])
	}
	if res.Summary.Added != 1 || res.Summary.Removed != 1 || res.Summary.Modified != 1 {
		t.Errorf("summary mismatch: added=%d removed=%d modified=%d",
			res.Summary.Added, res.Summary.Removed, res.Summary.Modified)
	}
}

func containsStr(list []string, s string) bool {
	for _, v := range list {
		if v == s || strings.HasSuffix(v, "."+s) {
			return true
		}
	}
	return false
}
