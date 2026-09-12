// Package closepolicy 提供跨 main 包与 Wails OnBeforeClose 回调之间共享的
// "关闭时是否最小化到托盘" 状态。
//
// 设计动机：Wails 在 wails.Run 期间构造 OnBeforeClose 闭包，而 config.Config
// 直到 app.startup 钩子触发后才完成初始化。把策略状态封装为原子布尔，
// 避免 main.go 直接依赖 config 包造成循环引用。
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

// 退出链路：先 SetForce(true)，再 Quit()（前端菜单与托盘退出共用）。
func SetForce(value bool) { forced.Store(value) }

// IsForced 供 main.go OnBeforeClose 读取；true 时即便 enabled=true 也放行。
func IsForced() bool { return forced.Load() }
