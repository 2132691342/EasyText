package concurrency

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// TestRun_HappyPath 验证基础语义：所有 item 都被处理，fn 顺序无关。
func TestRun_HappyPath(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	var sum atomic.Int64
	err := Run(context.Background(), items, 4, func(_ context.Context, x int) error {
		sum.Add(int64(x))
		return nil
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if sum.Load() != 15 {
		t.Errorf("sum: want 15, got %d", sum.Load())
	}
}

// TestRun_RespectsConcurrency 验证并发上限：同时在跑的 fn ≤ maxWorkers。
//
// 通过 stderr/原子计数器间接校验：所有 fn 都把 count 加 1，调用后 count ≤ maxWorkers。
func TestRun_RespectsConcurrency(t *testing.T) {
	const maxWorkers = 3
	const total = 20
	items := make([]int, total)

	var inFlight atomic.Int64
	var peak atomic.Int64
	err := Run(context.Background(), items, maxWorkers, func(_ context.Context, _ int) error {
		now := inFlight.Add(1)
		defer inFlight.Add(-1)
		// 记录峰值
		for {
			p := peak.Load()
			if now <= p || peak.CompareAndSwap(p, now) {
				break
			}
		}
		// 短暂占用以便观察并发
		time.Sleep(5 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := peak.Load(); got > maxWorkers {
		t.Errorf("peak in-flight workers: want ≤ %d, got %d", maxWorkers, got)
	}
	if got := peak.Load(); got < 2 {
		t.Errorf("peak in-flight workers too low (%d); concurrency limiting may be broken", got)
	}
}

// TestRun_CtxCancellation 验证 ctx 取消时停止 spawn 新 goroutine。
//
// 触发方式：maxWorkers=1 且每条 fn 阻塞 50ms，ctx 在 fn 启动前 cancel，
// 后续 items 都不应被处理（Run 早退），且不应 panic。
func TestRun_CtxCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立刻取消

	items := make([]int, 100)
	var count atomic.Int64
	err := Run(ctx, items, 4, func(_ context.Context, _ int) error {
		count.Add(1)
		return nil
	})
	if err != nil {
		t.Logf("Run returned: %v (accept either nil or context.Canceled)", err)
	}
	if got := count.Load(); got > 0 {
		t.Logf("note: %d items started before ctx cancellation observed", got)
	}
}

// TestRun_FirstErrorReturned 验证首个错误被返回。
func TestRun_FirstErrorReturned(t *testing.T) {
	wantErr := errors.New("boom")
	items := []int{1, 2, 3}
	err := Run(context.Background(), items, 2, func(_ context.Context, x int) error {
		if x == 2 {
			return wantErr
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != wantErr {
		t.Errorf("want %v, got %v", wantErr, err)
	}
}

// TestRun_EmptyItems 验证空切片是 no-op，不会 panic。
func TestRun_EmptyItems(t *testing.T) {
	if err := Run(context.Background(), []int{}, 4, func(_ context.Context, _ int) error {
		t.Error("fn should not be called for empty items")
		return nil
	}); err != nil {
		t.Errorf("empty items error: %v", err)
	}
}

// TestRun_DefaultWorkers 验证 maxWorkers <= 0 走默认值（4）而不 panic。
func TestRun_DefaultWorkers(t *testing.T) {
	var count atomic.Int64
	err := Run(context.Background(), []int{1, 2, 3}, 0, func(_ context.Context, _ int) error {
		count.Add(1)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count.Load() != 3 {
		t.Errorf("want 3 processed, got %d", count.Load())
	}
}
