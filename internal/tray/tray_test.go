package tray

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestWaitForExit_ImmediatelyReturnsWhenNotStarted 验证未 Start 时，WaitForExit 立即返回。
// 修复后没有 sentinel 会导致 WaitForExit 永久阻塞，导致进程无法退出的潜在 bug。
func TestWaitForExit_ImmediatelyReturnsWhenNotStarted(t *testing.T) {
	// 由于其他测试可能动过状态，重置 done
	doneOnce = sync.Once{}
	done = closedChan()

	start := time.Now()
	ok := WaitForExit(5000) // 给 5s 超时，未启动状态下应立即返回
	elapsed := time.Since(start)

	if !ok {
		t.Error("WaitForExit should return true when not started (done already closed)")
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("WaitForExit should be near-instant when done is closed, took %v", elapsed)
	}
}

// TestQuit_Idempotent 验证 Quit 多次调用安全：未启动是 no-op，不 panic。
func TestQuit_Idempotent(t *testing.T) {
	// 重置为未启动
	mu.Lock()
	started = false
	mu.Unlock()
	doneOnce = sync.Once{}
	done = closedChan()

	// 不应 panic，也不应阻塞
	Quit()
	Quit()
	Quit()
}

// TestSetHandlers_ThreadSafe 验证并发 Set/Read handlers 不 race。
// 配 -race 时这条测试必须通过，否则后台 goroutine 触发 Quit 时可能读 nil handler。
func TestSetHandlers_ThreadSafe(t *testing.T) {
	// 重置
	mu.Lock()
	showHandler = nil
	quitHandler = nil
	started = false
	mu.Unlock()

	calls := atomic.Int64{}

	go func() {
		for i := 0; i < 1000; i++ {
			SetHandlers(
				func() { calls.Add(1) },
				func() { calls.Add(1) },
			)
		}
	}()
	for i := 0; i < 1000; i++ {
		mu.RLock()
		_ = showHandler
		_ = quitHandler
		mu.RUnlock()
	}
}
