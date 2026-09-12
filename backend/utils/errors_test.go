package utils

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
)

// TestWrapError_PreservesChain 验证 WrapError 保留根因链：
// errors.Is 可穿透到底层错误，Error() 展示业务消息。
func TestWrapError_PreservesChain(t *testing.T) {
	wrapped := WrapError(1001, "无法读取文件", fs.ErrNotExist)

	if !errors.Is(wrapped, fs.ErrNotExist) {
		t.Errorf("errors.Is should match root fs.ErrNotExist, got: %v", wrapped)
	}
	if got := wrapped.Error(); got == "" || !strings.Contains(got, "无法读取文件") {
		t.Errorf("Error() should contain message, got: %q", got)
	}
}
