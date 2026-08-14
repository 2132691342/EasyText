// Package tray 提供最小化的系统托盘实现，供 main 接入。
//
// 设计要点：
//   - 单一全局实例（Start 只能调一次）；Start/Quit 控制生命周期。
//   - 调用方通过 SetHandlers 注入 ShowFn / QuitFn，供托盘菜单点击时调用。
//   - 图标与 tooltip 通过 embed 注入，避免运行时 I/O。
//
// Wails 集成：
//   - main.go 通过 OnBeforeClose 拦截关闭按钮，根据 closepolicy 状态决定是否隐藏窗口。
//   - 用户从托盘「退出」菜单触发：QuitHandler → runtime.Quit(ctx) → Wails OnShutdown → 本包的 Quit()，
//     形成完整链路，进程退出前会让 systray 干净地 Shell_NotifyIcon(NIM_DELETE)。
//   - Quit 幂等；WaitForExit 提供带超时的退出同步，避免进程 exit 后托盘图标短暂残留。
package tray

import (
	_ "embed"
	"runtime"
	"sync"
	"time"

	"github.com/getlantern/systray"
)

//go:embed icon.ico
var iconData []byte

// ShowFn/QuitFn 是托盘菜单点击时的回调，由调用方在 SetHandlers 中注入。
type ShowFn func()
type QuitFn func()

var (
	mu          sync.RWMutex
	showHandler ShowFn
	quitHandler QuitFn
	// done 由 start goroutine 在内部 wrapper 退出时关闭，作为 WaitForExit 的同步信号。
	// 初始为已关闭状态——启动前 WaitForExit 立即返回，避免语义混乱。
	doneOnce sync.Once
	done     = closedChan()

	// 文案/状态
	iconTitle = "EasyText"
	tooltip   = "EasyText — 轻量级文本编辑器"
	menuShow  = "显示主窗口"
	menuQuit  = "退出"

	// 启动幂等性：Start 多次调用只生效一次。
	started bool
)

func closedChan() chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

// SetHandlers 注册点击「显示主窗口」/「退出」菜单项时的回调。
// 必须在 Start 之前调用一次。
func SetHandlers(show ShowFn, quit QuitFn) {
	mu.Lock()
	defer mu.Unlock()
	showHandler = show
	quitHandler = quit
}

// SetTexts 调整菜单文案（用于 i18n）。
func SetTexts(title, tip, showLabel, quitLabel string) {
	if title != "" {
		iconTitle = title
	}
	if tip != "" {
		tooltip = tip
	}
	if showLabel != "" {
		menuShow = showLabel
	}
	if quitLabel != "" {
		menuQuit = quitLabel
	}
}

// Start 在独立 goroutine 启动 systray；非阻塞。重复调用返回 false。
//
// onReady 由 systray 在消息循环就绪后调用，负责构建菜单。
// onExit 由 systray 在退出时调用，用于关 done 让 WaitForExit 返回。
func Start(onReady func(), onExit func()) bool {
	mu.Lock()
	if started {
		mu.Unlock()
		return false
	}
	started = true
	mu.Unlock()

	// 重置 done：未被启动时默认已关闭，启动后改为未关闭、由 onExit 关闭。
	doneOnce = sync.Once{}
	done = make(chan struct{})

	go func() {
		// Windows 上 systray 的窗口创建与消息循环必须在同一 OS 线程：
		// 锁定本 goroutine 到当前线程，避免 GetMessage 与 CreateWindowEx
		// 分属不同线程导致托盘菜单点击/退出回调收不到消息。
		runtime.LockOSThread()
		wrappedExit := func() {
			defer closeOnce(&doneOnce, done)
			if onExit != nil {
				onExit()
			}
		}
		systray.Run(onReady, wrappedExit)
	}()
	return true
}

// closeOnce 让 wrappedExit 即使重复执行也只关闭一次 done 通道。
// sync.Once 在多个 goroutine 同时触发也能保证幂等。
func closeOnce(o *sync.Once, ch chan struct{}) {
	o.Do(func() { close(ch) })
}

// Quit 主动停止 systray 消息循环。幂等——未启动时也是 no-op。
//
// 调用后 systray goroutine 会让消息循环退出，触发 wrappedExit → close(done)。
func Quit() {
	mu.RLock()
	s := started
	mu.RUnlock()
	if !s {
		return
	}
	systray.Quit()
}

// WaitForExit 阻塞至 systray goroutine 退出或超时（毫秒）。
//
// 用于 main() 在 Quit 之后给 systray 一个清理窗口，
// 避免进程 exit 后 Windows 资源短暂泄漏（例如托盘图标残留）。
// 超时返回 false 不视为错误——主进程即将退出。
func WaitForExit(timeoutMs int) bool {
	if timeoutMs <= 0 {
		timeoutMs = 200
	}
	t := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
	defer t.Stop()
	select {
	case <-done:
		return true
	case <-t.C:
		return false
	}
}

// OnReady 由 Start 的 onReady 参数传入：构建菜单并绑定图标。
func OnReady() {
	systray.SetIcon(iconData)
	systray.SetTitle(iconTitle)
	systray.SetTooltip(tooltip)
	mShow := systray.AddMenuItem(menuShow, "显示 EasyText 主窗口")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem(menuQuit, "退出 EasyText")

	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				mu.RLock()
				fn := showHandler
				mu.RUnlock()
				if fn != nil {
					fn()
				}
			case <-mQuit.ClickedCh:
				mu.RLock()
				fn := quitHandler
				mu.RUnlock()
				if fn != nil {
					fn()
				}
				return
			}
		}
	}()
}
