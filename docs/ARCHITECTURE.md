# EasyText 架构与实现

EasyText 是一款 Windows 桌面文本编辑器，面向开发者的日常文档与代码编辑场景。
本文描述项目的整体设计、每个模块的职责与关键实现，是理解代码的入口。

## 1. 技术栈

| 层级 | 技术 | 选型理由 |
|------|------|----------|
| 桌面框架 | Wails v2.12（Go + 系统 WebView2） | 单 EXE、体积小、无需打包浏览器内核 |
| 后端 | Go 1.24 | 静态编译、并发模型简单 |
| 前端框架 | Vue 3 + TypeScript + Composition API | 组件化 + 类型安全 |
| 状态管理 | Pinia | 轻量 store |
| UI 基建 | Element Plus（表单/消息）+ Tailwind CSS + Lucide 图标 | 表单效率 + 原子样式 + 统一图标 |
| 编辑器 | CodeMirror 6 | 模块化扩展体系、60+ 语言支持 |
| Markdown | markdown-it + Mermaid + KaTeX | 预览、图表、公式 |
| 存储 | GORM + glebarez/sqlite（纯 Go，无 CGO） | 交叉编译友好 |
| 编码 | chardet + golang.org/x/text | 37 种编码检测与转换 |
| 脚本 | gopher-lua | 内嵌脚本引擎，超时保护 |
| 文件监控 | fsnotify | 实时文件变化 |

## 2. 总体架构

```
┌─────────────────────────── Windows 进程 ───────────────────────────┐
│                                                                     │
│  main.go          生命周期：单实例检测 → Wails 应用 → 托盘           │
│  app.go           App{ embed *api.Handler }，仅透传，无业务逻辑      │
│                                                                     │
│  ┌── WebView2（frontend/dist 内嵌）──────────────────────────────┐  │
│  │  Vue 3 组件树 ←→ wailsjs 绑定（116 个方法）                    │  │
│  └──────────────┬────────────────────────────────────────────────┘  │
│                 │ 绑定调用 / 事件（EventsOn / EventsEmit）            │
│  ┌──────────────▼────────────────────────────────────────────────┐  │
│  │ backend/api   绑定层：按特性分文件的薄方法，校验后委托          │  │
│  ├───────────────────────────────────────────────────────────────┤  │
│  │ backend/tools  业务服务：json/diff/encoding/findreplace/…      │  │
│  │ backend/file   文件域：reader/writer/tree/watcher              │  │
│  │ backend/config 配置 + SQLite（GORM）                           │  │
│  │ backend/tray / closepolicy / singleinstance / fileassoc        │  │
│  │ backend/concurrency / utils（AppError + logger）               │  │
│  └───────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

分层规则：

1. **Wails 绑定层**（`backend/api`）：所有导出方法挂在 `Handler` 及其嵌入的子
   handler 上，前端经 `App` 的方法提升直接调用。只做参数处理与委托，不含业务。
2. **业务服务层**（`backend/tools`、`backend/file`）：纯逻辑，可脱离 Wails 测试。
3. **基础设施层**（`backend/config`、`backend/utils` 等）：数据库、配置、错误、日志。

## 3. 目录结构

```
main.go / app.go              Wails 入口；App 嵌入 *api.Handler
wails.json                    应用信息与构建配置
backend/
  api/                        绑定层：handler.go（装配/Startup）+ api_*.go 按特性分文件
                              handler_subsystem.go 子 handler（Recent/FileAssoc/Search）
  tools/                      业务服务：bookmark compare convert diff draft encoding
                              findreplace json macro recent regex script snippet
  file/                       reader（分块读取）writer（原子写）tree（目录树）watcher
  config/                     AppConfig（v3 schema）+ Source 接口 + GORM SQLite
  closepolicy/                「关闭到托盘」与强制退出标志（原子布尔）
  singleinstance/             命名互斥 + TCP IPC（二次拉起转发文件）
  fileassoc/                  Windows 注册表文件关联（HKCU，不抢默认）
  tray/                       托盘生命周期（Start/Quit/WaitForExit）
  concurrency/                有界并行执行器 Run
  utils/                      AppError（错误码分段）+ logger
frontend/
  src/
    components/
      MainLayout.vue          顶层装配 + 全局事件接线
      NddMenuBar.vue          菜单栏（28px，键盘可达，超长列表内部滚动）
      NddToolbar.vue          工具栏（30px，Sublime 风即时 Popover）
      TabBar.vue              标签栏（30px）
      StatusBar.vue           状态栏（24px）
      Sidebar.vue             侧边栏（文件树/面板容器）
      ModalOverlay.vue        所有浮窗统一容器
      editor/CodeEditor.vue   编辑器容器（创建/换绑/命令分发）
      editor/composables/     主题/语言/补全/书签/列编辑/宏/MD 七个扩展模块
      editor/ext/             keymap 装配 + 右键菜单数据
      viewer/                 Hex/图片/日志/取色器查看器
      *.vue                   各功能浮窗（Diff/批量替换/编码转换/正则测试…）
    composables/              useCommands（命令分发）/ useFileOps / useTailWatcher
    stores/                   Pinia：editor / file / setting / converterTab
    style.css                 设计令牌 --et-*（颜色/间距/字号/高度阶梯/语义色）
    scripts/gen-bindings.cjs  从 Go 源码生成 wailsjs 绑定
    wailsjs/                  绑定产物（已提交，前端 typecheck 不依赖 Wails CLI）
```

## 4. 后端设计

### 4.1 启动流程

```
main.go
  ├─ singleinstance.New()          已有实例 → 转发文件参数后退出
  ├─ wails.Run(...)                窗口 / 资产 / 生命周期钩子
  │    OnStartup  → app.startup → handler.Startup(ctx)
  │    OnBeforeClose → closepolicy 决定拦截（隐藏到托盘）或放行
  │    OnShutdown → 清理
  └─ tray.Start(...)               托盘驻留
```

`Handler.Startup`（backend/api/handler.go）：

1. 初始化日志与配置目录；
2. `config.InitDatabase`——失败直接 panic（fail-fast：SQLite 支撑草稿/书签/
   片段/最近文件，起不来不如早死）；
3. `config.InitConfig`（非致命，损坏时回落默认配置）；
4. 构造 `cfgSource`（`config.Source` 接口）并注入全部 DB-依赖服务；
5. Startup 之后服务保证非 nil，**绑定层不做 nil 守卫**。

### 4.2 配置与存储（backend/config）

- `AppConfig`（v3 schema）按段拆分：Editor / Theme / File / UI / ColumnMode /
  Shortcut。`mergeConfig` 用「读到的值覆盖默认值」合并，旧版本配置升级时新字段
  自然回落默认值；版本号写回 3。
- 业务服务不直读全局配置，而是依赖 `Source` 接口取值（可注入、可测试）。
- SQLite 表：`Setting`、`Draft`（未保存草稿）、`Snippet`（代码片段）、
  `Bookmark`（全局书签）、`RecentEntry`（最近文件/文件夹）、`RemoteConnection`。

### 4.3 关键业务服务（backend/tools）

| 服务 | 职责与要点 |
|------|-----------|
| findreplace | 单文件查找/替换（大小写/全词/正则），目录级并发查找与批量替换，ctx 取消即停 |
| encoding | 37 种编码互转；BOM 识别覆盖 UTF-8/16LE/16BE/32LE/32BE 四族歧义 |
| json | 格式化/压缩/校验（带行号）/JSONPath 查询/结构体生成/键提取/结构化 Diff |
| diff & compare | LCS 行级 Diff + 字符级高亮；文件、目录对比 |
| convert | JSON/YAML/TOML/XML 互转 |
| draft | 按文件路径自动保存草稿；CheckConflict 三分支判定磁盘/草稿新旧 |
| bookmark | 全局书签；同文件同行去重，空 note 不覆盖旧值 |
| snippet | 代码片段；支持 VS Code JSON 与 EasyText 数组格式互导 |
| recent | 最近文件/文件夹；重复打开去重，超上限滚动淘汰 |
| script | gopher-lua 执行器；超时（5s）返回失败结果；Lua VM 由执行 goroutine 独占创建并 Close，杜绝双重 Close panic |
| macro | 宏录制（前端持久化为主，Go 侧仅保留兼容接口） |
| regex | 正则测试（匹配 + 捕获组提取） |

### 4.4 文件域（backend/file）

- **reader**：大文件分块读取 `ReadPartial(path, offset, count)`；offset 越界返回
  空串，作为前端增量加载的停止条件。
- **writer**：保存时原子写（临时文件 + 替换），保留换行风格与编码。
- **tree**：懒加载目录树，忽略隐藏目录。
- **watcher**：fsnotify 监听活动文件；外部修改事件经前端确认后重载（见 5.4）。

### 4.5 进程级机制

- **退出策略（closepolicy）**：两个原子标志 `enabled`（用户是否开启「关闭到
  托盘」）与 `forced`（强制退出）。`OnBeforeClose` 检查：`IsForced() || 未开启`
  → 放行退出；否则拦截并隐藏窗口。前端菜单退出必须走
  `SetCloseToTray(false) → ForceQuit() → Quit()` 链路，否则被拦截。
- **单实例（singleinstance）**：命名互斥保证唯一；二次启动通过本机 TCP 把
  命令行中的文件路径转发给运行实例后退出。
- **托盘（tray）**：`Start/Quit/WaitForExit`；`done` channel + `sync.Once`
  保证未启动时 Wait 立即返回、Quit 幂等。
- **文件关联（fileassoc）**：HKCU 下注册 ProgID `EasyText.text`，只出现在
  「打开方式」列表，不抢占默认程序；安装器与便携注册共用同一 ProgID。
- **并发执行器（concurrency.Run）**：有界 worker 池，返回首个错误；目录查找/
  批量替换/脚本并发均构建其上。
- **错误体系（utils）**：`AppError{Code, Message, Cause}`，`WrapError` 保留
  `errors.Is/As` 链。错误码分段：1xxx 文件、2xxx JSON、3xxx 编码、4xxx diff、
  5xxx 配置/db、6xxx compare、8xxx draft/bookmark、9xxx script。

## 5. 前端设计

### 5.1 Chrome 五层体系

窗口框架由五层固定高度的横向条带构成，全部基于 `style.css` 中的 `.et-chrome-*`
工具类与 `--et-*` 令牌：

| 层 | 组件 | 高度 |
|----|------|------|
| 菜单栏 | NddMenuBar | 28px |
| 工具栏 | NddToolbar | 30px |
| 标签栏 | TabBar | 30px |
| 侧边栏 | Sidebar | 弹性宽度 |
| 状态栏 | StatusBar | 24px |

约定：新 UI 只允许使用 `--et-*` 变量（颜色/间距/字号/圆点/语义色），禁止新增
字面色值；所有浮窗（设置、对比、批量替换、正则测试……）一律包在
`ModalOverlay` 里——它统一处理 Teleport、遮罩、focus-trap、Esc 关闭、滚动锁
与过渡动画，组件自身只写内容。

### 5.2 状态管理

四个 Pinia store：

- `editorStore`：标签页数组、活动标签、脏标记、撤销/前进位置栈、宏状态。
- `fileStore`：目录树、当前目录。
- `settingStore`：配置镜像（主题/字号/最近文件上限等），变更即写后端。
- `converterTabStore`：格式转换器页签状态。

### 5.3 命令系统

`useCommands.ts` 是前端命令中枢：每个命令对应一个 id，统一处理快捷键、菜单、
工具栏的触发。CodeMirror 快捷键经 `editor/ext/ext-keymap.ts` 的 `CMD_ALIASES`
映射到同一套命令 id，保证「菜单点击」与「按键」走同一条路径。右键菜单项由
`ext/ext-context-menu.ts` 的纯数据驱动。

### 5.4 编辑器架构（components/editor）

```
CodeEditor.vue（容器，~1300 行）
  ├─ createEditor()：组装 CodeMirror View，各关注点一个 compartment，
  │   运行时切换（主题/换行/缩进/语言/折叠/空白显示）走 dispatch reconfigure，
  │   不销毁重建
  ├─ StateField：书签行（markField）、词高亮（highlightWordField）、
  │   外链预览（webAddrField）
  ├─ updateListener：内容变更 → editorStore.updateTabContent（防抖持久化）
  │   + 光标位置上报 + 宏步骤采集
  ├─ tab.id 变化 → 原地换绑文档（保存/恢复滚动位置与撤销历史指针）
  ├─ tab.content 外部变更 → 受控同步（tail 重载场景）
  └─ Minimap.vue：视口矩形 {scrollTop, scrollHeight, clientHeight}，
      编辑器滚动 rAF 同步
composables/（互不 import，经 getEditorView 回调拿 View 实例）
  useEditorTheme / Language / Completion / Bookmark / ColumnMode / Macro / Markdown
ext/
  ext-keymap.ts（CMD_ALIASES + buildPrecKeymap）
  ext-context-menu.ts（菜单数据 + 可见性状态）
```

要点：

- **就地换绑**：切换标签页用 `dispatch({changes})` 换文档，不销毁 View，
  保证撤销栈与滚动行为可控；滚动位置在换绑前后保存/恢复。
- **补全缓存失效**：用户增删片段后调用 `invalidateCompletionCache()`，
  下一次补全重新拉取。
- **书签同步**：DB 书签 200ms 防抖同步进 editorStore；F2 切换与面板展示
  消费同一份数据。

### 5.5 全局事件接线（MainLayout.vue）

- `EventsOn('file:change')`：后端 watcher → 前端 `checkExternalChanges`，
  由用户确认后重载；
- 活动标签变化 → `StartFileWatch / StopFileWatch`，只监听当前文件；
- 侧边栏宽度等 UI 状态持久化到 Setting 表。

### 5.6 绑定生成

`frontend/wailsjs/` 随 `wails build` / `wails dev` 自动重新生成，并**提交进
仓库**（备用手工生成脚本：`frontend/scripts/gen-bindings.cjs`）：前端
`vue-tsc` 类型检查不依赖 Wails CLI，CI 因此无需安装 Go 工具链即可构建前端。

## 6. 典型数据流

**打开并编辑文件**：`useFileOps.openFile` → `ReadFile` → editorStore 建标签 →
CodeEditor 就地换绑 → 输入 → updateListener 防抖 `updateTabContent` → 自动保存
（含 Draft 草稿兜底）→ `Ctrl+S` 时 `SaveFile`（保留编码与换行风格）。

**外部修改文件**：watcher 触发 `file:change` → MainLayout 调
`checkExternalChanges` → 弹出确认 → 重载内容并恢复光标。

**退出**：菜单退出 → `SetCloseToTray(false)` → `ForceQuit()`（置 forced 标志）
→ `Quit()` → `OnBeforeClose` 放行 → 进程退出；直接点窗口 ✕ 则被拦截，隐藏
到托盘。

## 7. 构建与发布

```bash
wails dev                                   # 开发：Vite HMR + Go 后端
wails build                                 # 产物 build/bin/EasyText.exe
wails build -platform windows/amd64 --nsis  # NSIS 安装包
```

NSIS 安装包：桌面/开始菜单快捷方式、文件关联勾选、WebView2 引导；卸载清理
程序目录与注册信息（build/windows/installer/project.nsi）。

## 8. 测试策略

测试聚焦**复杂流程与长链路**，不做逐字段的简单断言（23 个测试，全量 < 5s）：

- `tools/findreplace`：单文件选项矩阵 + 百文件目录查找/批量替换/ctx 取消；
- `tools/script`：超时拦截 + 并发执行（测试内把 `luaTimeout` 缩到 300ms）；
- `tools/encoding`：GBK 往返 + 四族 BOM 识别与剥离往返；
- `tools/{draft,bookmark,snippet,recent,macro}`：各自的生命周期与业务规则
  （草稿冲突三分支、书签同行去重、片段 VS Code 导入导出、最近文件去重与上限）；
- `api`：Startup 服务装配不变量；`config`：合并与 v2→v3 迁移；
- `file`：分块读取边界；`tray`：退出幂等；`utils`：错误链。

测试约定：

- DB 用例使用 `setupDBTest(t)`（TempDir + `InitDatabase`），结束时必须
  `closeDB(config.DB)`，否则 Windows 上临时目录删不掉；
- 带私有 `storagePath` 的服务（Macro）在同包测试中直接覆盖该字段并手动 `load()`。

```bash
cd backend && go test ./... -count=1 -timeout 120s
```

## 9. 约束与已知边界

- 应用为 Windows 专属：fileassoc/singleinstance 依赖 Win32 API，CI 的 backend
  job 在 windows runner 上执行；
- Wails 把 Go `[]byte` 序列化为 JS `number[]`，二进制接口需元素级转换；
- 单窗口模型（Wails v2）；
- goroutine 纪律：`defer Close` 只在创建者 goroutine 内执行，`wg.Add` 先于
  `go` 语句，goroutine 内 panic 统一 recover 兜底。
