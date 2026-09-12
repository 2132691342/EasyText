package tray

import (
	"sync"
	"testing"
	"time"
)

// TestTray_ExitSemantics 验证退出语义：未启动时 WaitForExit 立即返回
// （防 sentinel 缺失导致的永久阻塞）、Quit 重复调用安全不 panic。
func TestTray_ExitSemantics(t *testing.T) {
	doneOnce = sync.Once{}
	done = closedChan()
	started = false

	start := time.Now()
	if !WaitForExit(5000) {
		t.Error("WaitForExit should return true when not started")
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Errorf("WaitForExit should be near-instant when done is closed, took %v", elapsed)
	}

	Quit()
	Quit()
	Quit()
}
