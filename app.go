package main

import (
	"context"

	"easy-text/backend/api"
	"easy-text/backend/config"
	"easy-text/internal/closepolicy"
	"easy-text/internal/tray"
)

// App 是 Wails 绑定所需的顶层结构体。
// 通过嵌入 *api.Handler，所有 API 方法自动提升到 App 上，
// 保持 Wails 绑定（main.App）不变，而实际逻辑全部在 backend/api 包中。
type App struct {
	*api.Handler
}

// NewApp 创建 App 实例
func NewApp() *App {
	return &App{Handler: api.NewHandler()}
}

// startup 是 Wails 约定的生命周期方法（小写），委托给 Handler.Startup。
// 同时把 ctx 与 closepolicy 同步到 package 级变量，供 main.go 的托盘回调使用。
func (a *App) startup(ctx context.Context) {
	appCtx = ctx
	a.Handler.Startup(ctx)
	// 初次同步：避免 OnBeforeClose 在配置完成前读到错误的旧值。
	if cfg := config.Config; cfg != nil {
		closepolicy.Set(cfg.Get().UI.CloseToTray)
	}
}

// shutdown 是 Wails 约定的生命周期方法（小写），委托给 Handler.Shutdown。
//
// 关停顺序：先让 Handler 做它的事（DB 关闭、日志 flush），再调 tray.Quit()
// 让 systray 干净地 Shell_NotifyIcon(NIM_DELETE)；最后 WaitForExit 给
// systray goroutine 一个 ≤ 200ms 的清理窗口，避免 Windows 上托盘图标短暂残留。
// 超时不会阻塞——主进程即将退出，超时仅是 cleanup 不及时，不影响退出。
func (a *App) shutdown(ctx context.Context) {
	a.Handler.Shutdown(ctx)
	tray.Quit()
	tray.WaitForExit(200)
}
