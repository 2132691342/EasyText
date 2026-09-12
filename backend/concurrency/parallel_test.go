package concurrency

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// TestRun 并发执行器的核心语义：并发上限不被突破、全部 item 被处理、
// 首个错误被返回、ctx 取消后及时退出。
func TestRun(t *testing.T) {
	t.Run("bounded-workers", func(t *testing.T) {
		const maxWorkers = 3
		items := make([]int, 20)
		var inFlight, peak, done atomic.Int64

		err := Run(context.Background(), items, maxWorkers, func(_ context.Context, _ int) error {
			now := inFlight.Add(1)
			defer inFlight.Add(-1)
			for {
				p := peak.Load()
				if now <= p || peak.CompareAndSwap(p, now) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			done.Add(1)
			return nil
		})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if got := peak.Load(); got > maxWorkers {
			t.Errorf("peak in-flight workers: want ≤ %d, got %d", maxWorkers, got)
		}
		if got := peak.Load(); got < 2 {
			t.Errorf("peak too low (%d); concurrency limiting may be broken", got)
		}
		if done.Load() != int64(len(items)) {
			t.Errorf("want all %d items processed, got %d", len(items), done.Load())
		}
	})

	t.Run("first-error", func(t *testing.T) {
		wantErr := errors.New("boom")
		err := Run(context.Background(), []int{1, 2, 3}, 2, func(_ context.Context, x int) error {
			if x == 2 {
				return wantErr
			}
			return nil
		})
		if err != wantErr {
			t.Fatalf("want %v, got %v", wantErr, err)
		}
	})

	t.Run("ctx-cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		done := make(chan struct{})
		go func() {
			defer close(done)
			_ = Run(ctx, make([]int, 100), 4, func(_ context.Context, _ int) error {
				return nil
			})
		}()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("Run did not return within 2s after ctx cancel")
		}
	})
}
