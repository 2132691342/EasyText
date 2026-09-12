# CLAUDE.md — EasyText 项目指南

> 给 AI 助手与接手开发者的项目速查。v2.1 重塑（M1–M4）后的架构基线。

## 常用命令

```bash
# 开发（需要 Windows 桌面环境 + WebView2）
wails dev                       # Vite HMR + Go 后端；自动生成 frontend/wailsjs/

# 后端（Linux 下也能跑，是 CI 的硬门禁）
cd backend && go build ./... && go vet ./...
cd backend && go test ./... -count=1 -timeout 120s

# 前端（wailsjs 未生成时 typecheck 会报 TS2307，先跑一次 wails dev 或用 CI 前置 wails build）
cd frontend && npm install && npm run build

# 一键 CI 等价检查
make ci-build
```

## 目录结构与分层

```
main.go / app.go            Wails 入口。禁止放业务逻辑；App 嵌入 *api.Handler
backend/
  api/                      Wails 绑定层（薄）：handler.go + api_*.go 按特性分文件
                            handler_subsystem.go 子 handler（Recent/FileAssoc/Search）
  file/                     文件域：reader / writer / tree / watcher
  tools/                    纯业务服务：json / diff / encoding / findreplace /
                            draft / snippet / bookmark / recent / script / macro / convert
  config/                   AppConfig（v3 schema）+ GORM SQLite（Draft/Snippet/
                            Bookmark/RecentEntry/RemoteConnection/Setting 表）
  closepolicy/              「关闭到托盘」+ force 退出标志（原子布尔）
  singleinstance/           命名互斥 + TCP IPC（文件关联二次拉起）
  concurrency/              通用有界并行执行器
  utils/                    AppError（错误码分段 1xxx~9xxx）+ logger
frontend/src/
  components/
    NddMenuBar / NddToolbar / TabBar / StatusBar / Sidebar   Chrome 五层（28/30/30/24px）
    ModalOverlay.vue         所有浮窗的统一容器（size/focus-trap/Esc/scroll-lock）
    MainLayout.vue           顶层装配 + 全局事件接线
    editor/
      CodeEditor.vue         容器（~1100 行）：createEditor / dispatch / 模板
      composables/           useEditorTheme / Language / Completion / Bookmark /
                             ColumnMode / Macro / Markdown（各司其职，互不 import）
      ext/                   ext-keymap（CMD_ALIASES + keymap）/ ext-context-menu
  composables/               useCommands（150+ 命令分发）/ useFileOps / useTailWatcher
  stores/                    Pinia：editor / file / setting / converterTab
  style.css                  设计令牌 --et-*（颜色/间距/字号/高度阶梯/语义色）
docs/bug-repro/             各里程碑的 bug 复现与验收剧本
```

## 核心约定（违反会回归）

1. **错误处理**：一律用 `utils.WrapError(code, msg, cause)` 保留 `errors.Is/As` 链。
   错误码分段：1xxx 文件、2xxx JSON、3xxx 编码、4xxx diff、5xxx 配置/db、
   6xxx compare、8xxx draft/bookmark、9xxx script/git。
2. **DB fail-fast**：`Handler.Startup` 中 DB 初始化失败直接 panic。
   Startup 之后 draft/snippet/bookmark/recent service 保证非 nil，**API 层不加 nil 守卫**。
3. **修复 bug 必须先写复现测试**（见 `encoding_test.go` 抓出 RemoveBOM 缺 UTF-32BE 分支的实例）。
4. **Goroutine 资源独占**：`defer L.Close()` 只能在执行 goroutine 内；
   goroutine 内 panic 用 `defer recover()` 兜底；`wg.Add` 先于 `go` 语句。
5. **Wails []byte 序列化**：Go `[]byte` → JS `number[]`。二进制接口需元素级转换或 base64。
6. **前端令牌**：新代码只用 `--et-*` 变量，禁止新增字面色（`#fff` / `bg-gray-*`）。
   浮窗一律包 `<ModalOverlay>`，不自写 Teleport+backdrop。
7. **编辑器扩展**：CodeMirror 功能进 `editor/composables/`（互不 import，接收
   `getEditorView` 回调）；命令别名进 `ext/ext-keymap.ts` 的 `CMD_ALIASES`；
   右键菜单数据进 `ext/ext-context-menu.ts`。
8. **退出链路**：前端退出必须 `SetCloseToTray(false)` → `ForceQuit()` → `Quit()`
   （closepolicy.IsForced 让 OnBeforeClose 放行），否则被「关闭到托盘」拦截。
9. **宏**：前端 editorStore（localStorage）是唯一事实源；Go 侧 api_macro 已标记
   Deprecated，不要把新功能挂上去。

## 测试

- 84 个用例 / 11 包。DB 相关测试用 `setupDBTest(t)`（TempDir + InitDatabase +
  `closeDB(config.DB)`，Windows 上必须关句柄否则 TempDir 删不掉）。
- `MacroService` 等带私有 storagePath 的服务：同包测试直接覆盖 `svc.storagePath`
  指向 TempDir，构造后需手动 `svc.load()`（构造函数里的 load 用的是用户目录）。
- 回归剧本：`docs/bug-repro/M3.md`。

## 已知边界（勿当 bug 重复修）

- `vue-tsc` 在 wailsjs 未生成时报 TS2307 —— 基线行为，跑 `wails dev` 即消。
- i18n（lang-en）只写配置不翻译模板字符串，是有意排除（见 M1–M3 plan §范围）。
- `new-window` 单窗口是 Wails v2 模型限制。
- `api_hash.go` 的 `ComputeFileHash` 是已知死 stub，无前端调用方。
