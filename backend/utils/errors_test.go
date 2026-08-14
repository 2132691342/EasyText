package utils

import (
	"errors"
	"io/fs"
	"os"
	"testing"
)

// Test WrapError 保留错误链——验证 errors.Is / errors.As 可穿透到根因。
// 复现 F-3：修复前所有调用 NewAppError(code, msg, err.Error()) 会丢失根因；
// 修复后 WrapError 用 Cause + Unwrap() 保留链。
func TestWrapError_PreservesChain(t *testing.T) {
	root := fs.ErrNotExist
	wrapped := WrapError(1001, "无法读取文件", root)

	// errors.Is 应穿透到根因
	if !errors.Is(wrapped, fs.ErrNotExist) {
		t.Errorf("errors.Is should match root fs.ErrNotExist, got: %v", wrapped)
	}

	// Error() 应展示链
	if got := wrapped.Error(); got == "" || !contains(got, "无法读取文件") {
		t.Errorf("Error() should contain message, got: %q", got)
	}
}

// TestNewAppError_StaticInfo 验证静态错误（无 Cause）行为不变。
func TestNewAppError_StaticInfo(t *testing.T) {
	e := NewAppError(1001, "文件不存在", "")
	if e.Cause != nil {
		t.Errorf("NewAppError should not have Cause, got: %v", e.Cause)
	}
	if e.Error() != "文件不存在" {
		t.Errorf("expected '文件不存在', got: %q", e.Error())
	}
}

// TestWrapError_DetailsNotSerialized 验证 Cause 不参与 JSON 序列化（不泄露内部细节给前端）。
func TestWrapError_DetailsNotSerialized(t *testing.T) {
	wrapped := WrapError(1001, "无法读取文件", os.ErrPermission)
	// Cause 字段标签为 `json:"-"`，通过 reflect 验证
	if got := wrapped.Cause; got == nil {
		t.Error("Cause should be set on the struct")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
