package main

import (
	"context"
	"embed"

	"easy-text/internal/closepolicy"
	"easy-text/internal/tray"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	// 注入关闭策略处理器：托盘「显示主窗口」时回显主窗口；
	// 「退出」时调用 runtime.Quit 真正退出进程。
	tray.SetHandlers(
		showMainWindow,
		quitApp,
	)
	tray.Start(tray.OnReady, nil)

	err := wails.Run(&options.App{
		Title:             "EasyText",
		Width:             1280,
		Height:            800,
		MinWidth:          800,
		MinHeight:         600,
		DisableResize:     false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		// 🆕 拦截关闭按钮：策略开启则隐藏并取消关闭；策略关闭则允许退出。
		OnBeforeClose: onBeforeClose,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
			WebviewUserDataPath:               "",
			WebviewBrowserPath:                "",
			Theme:                             windows.SystemDefault,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// onBeforeClose 由 Wails 在收到窗口关闭请求时回调。
// 返回 prevent=true 表示阻止关闭，让窗口隐藏到托盘（依赖 closepolicy 状态）。
func onBeforeClose(ctx context.Context) (prevent bool) {
	if !closepolicy.IsEnabled() {
		return false
	}
	runtime.WindowHide(ctx)
	return true
}

// showMainWindow 把窗口唤到前台（由托盘菜单点击触发）。
func showMainWindow() {
	if appCtx == nil {
		return
	}
	runtime.WindowShow(appCtx)
	runtime.WindowUnminimise(appCtx)
}

// quitApp 真正退出进程。
//
// runtime.Quit 内部会先调用 OnBeforeClose，而 OnBeforeClose 在「关闭到托盘」
// 开启时返回 true 会拦截退出（直接 return，进程不退出）。这里先禁用关闭到托盘，
// 让 OnBeforeClose 放行，才能走完 winc.Exit 的退出流程。
func quitApp() {
	if appCtx != nil {
		closepolicy.Set(false)
		runtime.Quit(appCtx)
	}
}

// appCtx 由 startup 钩子写入，供托盘回调后续使用。
var appCtx context.Context
