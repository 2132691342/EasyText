// Package closepolicy 提供跨 main 包与 Wails OnBeforeClose 回调之间共享的
// "关闭时是否最小化到托盘" 状态。
//
// 设计动机：Wails 在 wails.Run 期间构造 OnBeforeClose 闭包，而 config.Config
// 直到 app.startup 钩子触发后才完成初始化。把策略状态封装为原子布尔，
// 避免 main.go 直接依赖 config 包造成循环引用。
//
// v2.1（修 bug #5）：增加 force 标志，让"前端点击退出 / 托盘点退出"
// 可以先一步把策略置为 false，再触发 Quit，避免 OnBeforeClose 把退出
// 误解成"关闭到托盘"。
package closepolicy

import "sync/atomic"

var (
	enabled atomic.Bool
	forced  atomic.Bool
)

// Set 原子写入新策略值。Handler.Startup 启动时调用，UI 层切换时也调用。
func Set(value bool) { enabled.Store(value) }

// IsEnabled 供 main.go 的 OnBeforeClose 回调读取。
func IsEnabled() bool { return enabled.Load() }

// SetForce 标记"用户已明确要求退出"，下一次 OnBeforeClose 强制放行。
// main.go 的 OnBeforeClose 检测到 forced=true 时直接 return false（放行）。
//
// 用例：useCommands.exit 走 deps → 先调 SetForce(true) 再调 Set(false) 再 Quit()。
// 托盘 quitApp 同理（已在 main.go 实现）。
func SetForce(value bool) { forced.Store(value) }

// IsForced 供 main.go OnBeforeClose 读取；true 时即便 enabled=true 也放行。
func IsForced() bool { return forced.Load() }
