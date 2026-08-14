package fileassoc

import (
	"testing"

	"golang.org/x/sys/windows/registry"
)

// TestDefaultExtensions_NotEmpty 防御：扩展名列表为空时容易导致注册循环零次的情况。
// 即使后续策略调整为"按需注册"，这里也至少要保证有非空默认列表。
func TestDefaultExtensions_NotEmpty(t *testing.T) {
	if len(DefaultExtensions) == 0 {
		t.Fatal("DefaultExtensions is empty")
	}
	for _, e := range DefaultExtensions {
		if e == "" || e[0] == '.' {
			t.Errorf("extension %q should not be empty or have leading dot", e)
		}
	}
}

// TestProgIDStable 防御：ProgID 是与 NSIS 安装器交互的协议字符串。
// 若被改动，安装器注册的 ProgID 主键会与便携注册的 ProgID 主键路径不一致，导致
// 「卸载后快捷方式残留」或「HKCU ProgID 双写」之类问题。
func TestProgIDStable(t *testing.T) {
	const want = "EasyText.text"
	if progID != want {
		t.Errorf("progID changed: want %q, got %q", want, progID)
	}
}

// TestIsRegistered_FalseWhenNoKey 验证在没有 EasyText 注册表项时返回 false。
// 这条测试用 t.Setenv + 真实的 HKCU，但 key 由我们直接在测试运行前删除过。
// 注意：实际生产 HKCU 下若 EasyText 已注册，本测试返回 true，但不会引起 false-fail。
func TestIsRegistered_FalseWhenNoKey(t *testing.T) {
	// 主动删除（如果存在），让测试可重复
	_ = registry.DeleteKey(registry.CURRENT_USER, `Software\EasyText`)
	if IsRegistered() {
		t.Error("IsRegistered should be false after deleting Software\\EasyText")
	}
}
