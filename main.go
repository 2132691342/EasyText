package main

import (
	"context"
	"embed"
	"fmt"
	"os"

	"easy-text/internal/closepolicy"
	"easy-text/internal/singleinstance"
	"easy-text/internal/tray"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// pendingFilePath 存储通过文件关联传入的待打开文件路径
var pendingFilePath string

func main() {
	// 单实例检测
	si, err := singleinstance.New()
	if err != nil {
		// 已有实例在运行，将文件路径发送给已有实例后退出
		if err == singleinstance.ErrAlreadyRunning {
			filePath := getFilePathFromArgs()
			if filePath != "" {
				_ = singleinstance.SendFileToRunningInstance(filePath)
			}
			os.Exit(0)
		}
		// 其他错误，继续运行（不影响用户体验）
		fmt.Println("单实例检测失败:", err)
	} else {
		// 启动 IPC 监听器
		_ = si.StartListener()
		defer si.Stop()

		// 启动 goroutine 处理接收到的文件路径
		go func() {
			for path := range si.FileChannel() {
				if path != "" {
					handleReceivedFilePath(path)
				}
			}
		}()
	}

	// 获取命令行参数中的文件路径
	pendingFilePath = getFilePathFromArgs()

	app := NewApp()
	// 注入关闭策略处理器：托盘「显示主窗口」时回显主窗口；
	// 「退出」时调用 runtime.Quit 真正退出进程。
	tray.SetHandlers(
		showMainWindow,
		quitApp,
	)
	tray.Start(tray.OnReady, nil)

	err = wails.Run(&options.App{
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

// getFilePathFromArgs 从命令行参数中获取文件路径
// Windows 文件关联打开时会将文件路径作为参数传入
func getFilePathFromArgs() string {
	// 遍历参数，找到第一个看起来像文件路径的参数
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		// 跳过 Wails 内部的标志参数
		if len(arg) > 0 && arg[0] != '-' {
			// 检查是否是有效的文件路径
			if _, err := os.Stat(arg); err == nil {
				return arg
			}
			// 即使文件不存在也返回，让后续处理报错
			return arg
		}
	}
	return ""
}

// handleReceivedFilePath 处理从其他实例接收到的文件路径
func handleReceivedFilePath(path string) {
	if appCtx == nil {
		// 应用还未完全启动，保存到 pendingFilePath
		pendingFilePath = path
		return
	}
	// 通过事件通知前端打开文件
	runtime.EventsEmit(appCtx, "app:open-file", path)
	// 确保窗口可见
	runtime.WindowShow(appCtx)
	runtime.WindowUnminimise(appCtx)
}
