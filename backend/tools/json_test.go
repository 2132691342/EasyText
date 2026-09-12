package tools

import (
	"strings"
	"testing"
)

// TestJSON_Toolchain 按用户操作链路串联核心工具：格式化 → 压缩往返 →
// 校验（错误必须带行号）→ JSONPath 查询命中与未命中。
func TestJSON_Toolchain(t *testing.T) {
	jt := NewJSONTool()

	r := jt.Format(`{"a":1,"b":[2,3]}`, 2)
	if !r.Success || !strings.Contains(r.Content, "\n") || !strings.Contains(r.Content, `"a": 1`) {
		t.Fatalf("Format failed: %+v (%q)", r.Error, r.Content)
	}

	m := jt.Minify(r.Content)
	if !m.Success || strings.Contains(m.Content, "\n") {
		t.Fatalf("Minify failed: %+v (%q)", m.Error, m.Content)
	}

	bad := jt.Validate("{\n  \"a\": 1,\n  \"b\": \n}")
	if bad.Success || bad.Error == nil || bad.Error.Line <= 0 {
		t.Fatalf("Validate should fail with line number, got %+v", bad)
	}

	content := `{"store":{"book":[{"title":"a"},{"title":"b"}]}}`
	results, err := jt.QueryPath(content, "$.store.book[*].title")
	if err != nil || len(results) != 2 {
		t.Fatalf("QueryPath: err=%v hits=%d", err, len(results))
	}
	if results, _ = jt.QueryPath(content, "$.nothing.here"); len(results) != 0 {
		t.Errorf("miss should return 0 hits, got %d", len(results))
	}
}

// TestJSON_StructuredDiff 验证结构化 Diff 的 added/removed/modified 分类与汇总。
func TestJSON_StructuredDiff(t *testing.T) {
	jt := NewJSONTool()
	left := `{"a":1,"b":"keep","c":true}`
	right := `{"a":2,"b":"keep","d":"new"}`

	res, err := jt.StructuredDiff(left, right)
	if err != nil {
		t.Fatalf("StructuredDiff: %v", err)
	}
	kinds := map[string]string{}
	for _, e := range res.Entries {
		kinds[e.Path] = e.Type
	}
	if kinds["$.a"] != "modified" || kinds["$.c"] != "removed" || kinds["$.d"] != "added" {
		t.Errorf("kinds mismatch: %v", kinds)
	}
	if res.Summary.Added != 1 || res.Summary.Removed != 1 || res.Summary.Modified != 1 {
		t.Errorf("summary mismatch: %+v", res.Summary)
	}
}
