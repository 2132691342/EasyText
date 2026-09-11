// Package concurrency 提供受限并行的执行辅助，避免每个调用点都重复实现
// "channel semaphore + WaitGroup + ctx 取消"那一坨样板。
//
// 设计动机（ch04 §1 / ch12 §高可用 §1）：
//   - 批量 I/O 场景（findreplace / diff / file 目录遍历）必须限制并发，
//     否则 1k+ 的批量替换瞬间把 Wails WebView2 主线程拖到卡死。
//   - 取消必须可重入：用户在 UI 点取消 / 切 tab / 关对话框时，正在跑的
//     goroutine 要能收到信号尽早收尾，而不是继续打满 CPU。
//
// 与 sync.WaitGroup + sync.Mutex 手工实现的对照：
//   - 集中一处出错更容易保证一致性（统一 ctx 处理、统一错误收集）
//   - 不引入 x/sync/errgroup 依赖（项目当前 Go 1.24，x/sync v0.22 需要 Go 1.25）
package concurrency

import (
	"context"
	"sync"
)

// Run 在 maxWorkers 个 goroutine 上并发执行 fn(item)。
//
//   - ctx 取消时，不启动新的 goroutine；已经在跑的 fn 拿到 ctx 让其自行判断退出。
//     本函数不会强杀 fn，调用方应在 fn 内周期性检查 ctx.Done()。
//   - 任一 fn 返回非空 error，整体结果错误为该 error（首个失败）；
//     但不会打断其它 fn 的执行 —— 若需要严格 parallel 失败短路，
//     调用方应在 fn 内部自行处理 ctx。
//   - 本函数在所有 fn 返回后或 ctx 取消时返回。
//
// items 可以是任意 slice；fn 必须线程安全。
func Run[T any](ctx context.Context, items []T, maxWorkers int, fn func(ctx context.Context, item T) error) error {
	if maxWorkers <= 0 {
		maxWorkers = 4
	}
	if len(items) == 0 {
		return nil
	}

	sem := make(chan struct{}, maxWorkers)
	var (
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
	)

	for _, item := range items {
		// ctx 已取消：不再启动新 goroutine；早退路径
		if ctx.Err() != nil {
			break
		}
		item := item
		wg.Add(1)
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			wg.Done()
			return firstErr
		}

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			if err := fn(ctx, item); err != nil {
				once.Do(func() { firstErr = err })
			}
		}()
	}

	wg.Wait()
	return firstErr
}
