# EasyText

> 轻量级桌面文档编辑工具 · 仿 notepad-- 体验 · Wails + Vue 3 + TypeScript

**EasyText** 是一款面向开发者的 Windows 桌面编辑器。冷启动 < 1 秒、安装包 ~11MB，内置 JSON / Diff / 格式转换 / 编码转换 / 文件对比 / Markdown 预览 / Lua 脚本 / 文件监控等高频开发工具，开箱即用。

![Windows 10/11](https://img.shields.io/badge/platform-Windows%2010%2F11-blue) ![Wails v2.12](https://img.shields.io/badge/Wails-v2.12-green) ![Go 1.24](https://img.shields.io/badge/Go-1.24-00ADD8) ![Vue 3](https://img.shields.io/badge/Vue-3.4-42b883) ![License MIT](https://img.shields.io/badge/license-MIT-yellow)

## 下载安装

安装包（NSIS，~11MB）已随仓库发布，可直接使用：

- **安装包**：[`build/bin/EasyText-amd64-installer.exe`](build/bin/EasyText-amd64-installer.exe) — 双击安装，可选注册为常见文本类型的编辑器
- **免安装单文件**：[`build/bin/EasyText.exe`](build/bin/EasyText.exe) — 绿色运行，零依赖

> 安装包需要 Windows 10 1809+ / 11（系统已内置 WebView2）。

## 核心特性

- **轻量原生**：系统 WebView2 渲染，单 EXE ~24MB，零依赖，零插件
- **代码编辑**：CodeMirror 6 引擎，60+ 语法高亮、代码折叠、多光标、括号匹配
- **格式转换**：JSON / YAML / TOML / XML 互转；JSONPath 查询；JSON 转结构体（Go/TS/Java/Python 等）；JSON 结构化 Diff
- **开发者工具箱**：LCS 行级 Diff + 字符级高亮 + 统一视图、37 种编码检测与转换、批量查找/替换/重命名、MD5/SHA 哈希、列块模式（Alt+X）
- **Markdown**：编辑/分屏/预览三模式，Mermaid 图表 + KaTeX 数学公式 + 代码块语法高亮，一键导出 HTML
- **脚本系统**：基于 gopher-lua 的脚本管理器，支持超时保护与并发安全，可自动化常见编辑任务
- **文件监控**：实时 tail -f 文件变化并自动重载
- **会话恢复**：启动时自动恢复上次打开的文件、工作空间（workspace.etws）、光标位置
- **暗色模式**：19 套主题，支持自动跟随系统主题切换
- **可扩展侧栏**：代码片段面板、全局书签面板、函数列表、文件监控面板、日志查看模式
- **正则测试器**：独立浮层实时匹配测试，支持捕获组提取
- **系统托盘**：关闭到托盘常驻，可从托盘恢复窗口或彻底退出
- **文件关联**：安装器可选注册为常见文本类型的编辑器，出现在右键「打开方式」列表（不抢占已有默认）

## 界面预览

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ 文件(F) 编辑(E) 查找(S) 视图(V) 编码(N) 语言(L) 设置(T) 工具(O) 对比 关于  │ ← 菜单栏
├──────────────────────────────────────────────────────────────────────────────┤
│ [📄][📂][💾] │ [✂][📋][📌] │ [↶][↷] │ [🔍][🔄] │ [🔖][✏] │ [🔎+][🔎-] │ ← 工具栏
├──────────┬───────────────────────────────────────────────────────────────────┤
│ 📁 文件树│  [标签栏] tab1 × | tab2 × | tab3 ×                              │
│ 📄 tab1  │ ┌──────────────────────────────────────────────────────────────┐ │
│ 📄 tab2  │ │  CodeMirror 编辑器 / 浮层组件（Diff/JSON/Snippet 等）         │ │
│          │ └──────────────────────────────────────────────────────────────┘ │
├──────────┴───────────────────────────────────────────────────────────────────┤
│ 缩放:110% | 语言:xml | 行:8526 列:52 | Unix(LF) | UTF8 | C:\path\to\file    │← 状态栏
└──────────────────────────────────────────────────────────────────────────────┘
```

## 快速开始

### 环境要求

- **Go** 1.24+ · **Node.js** 18+ · **Wails CLI**（`go install github.com/wailsapp/wails/v2/cmd/wails@latest`）
- **Windows 10 1809+ / 11**（WebView2 已内置）

### 开发模式

```bash
wails dev                 # Vite HMR + Go 后端，自动重新生成 wailsjs 绑定
```

### 生产构建

```bash
wails build                                  # 输出 build/bin/easy-text.exe (~24MB)
wails build -platform windows/amd64 --nsis   # NSIS 安装包 (~11MB)
```

### 跨栈本地验证（无需 Wails 桌面环境）

```bash
make ci-build           # Linux 子集：backend build/vet/test + frontend typecheck/build
```

## 命令一览

| 命令 | 作用 |
|---|---|
| `make backend` | 后端 `go build` + `go vet` |
| `make frontend` | 前端 `npm ci` + `typecheck` + `build` |
| `make test` | 后端 `go test ./...`（52 个测试用例，60s 超时） |
| `make lint` | 前端 ESLint |
| `make ci-build` | CI 等价检查（Linux 子集） |
| `make clean` | 清理构建缓存 |

## 常用快捷键

| 快捷键 | 功能 | 快捷键 | 功能 |
|--------|------|--------|------|
| `Ctrl+O` | 打开文件 | `Ctrl+S` | 保存 |
| `Ctrl+F` / `Ctrl+H` | 查找 / 替换 | `Ctrl+G` | 转到行 |
| `Ctrl+Z` / `Ctrl+Y` | 撤销 / 重做 | `Ctrl+/` | 行注释 |
| `Ctrl+D` | 复制当前行 | `Ctrl+L` | 删除当前行 |
| `Ctrl+Shift+F` | 全局搜索 | `Ctrl+Shift+D` | 目录查找 |
| `Alt+X` | 列编辑模式 | `F11` / `Ctrl+P` | 全屏 / 打印 |
| `F2` | 切换书签 | `F3` / `Shift+F3` | 查找下一个/上一个 |

## 项目结构

```
EasyText/
├── main.go                       # Wails 入口（嵌入 frontend/dist + 绑定 main.App）
├── app.go                        # App 结构（embed *api.Handler），禁止放业务逻辑
├── wails.json                    # Wails 配置
├── go.mod / go.sum               # Go 依赖（easy-text 模块）
├── Makefile                      # 顶层开发命令（make ci-build 等）
├── .golangci.yml                 # Go 静态检查（errcheck / vet / staticcheck / gosec / revive）
├── CLAUDE.md                     # 给 Claude Code 的项目指南
├── backend/
│   ├── api/                      # Wails 绑定层（薄，每个特性一个文件）
│   │   ├── handler.go            # Handler 主结构 + Startup/Shutdown
│   │   ├── api_files.go          # 文件 I/O API
│   │   ├── api_diff.go           # 文档对比
│   │   ├── api_encoding.go       # 编码转换
│   │   ├── api_json.go           # JSON 工具
│   │   ├── api_compare.go        # 文件/目录对比
│   │   ├── api_script.go         # 脚本执行入口
│   │   └── ...                   # 其余按特性分文件
│   ├── file/                     # 文件域：reader / writer / tree / watcher
│   ├── tools/                    # 纯工具：json / diff / encoding / script / findreplace / regex ...
│   ├── config/                   # AppConfig（v3 schema）+ SQLite 数据库
│   └── utils/                    # AppError + global logger
├── frontend/
│   ├── package.json
│   ├── vite.config.ts            # Vite 配置（含 @ → src/ 别名）
│   ├── .eslintrc.cjs             # 前端 ESLint（no-explicit-any + no-empty）
│   └── src/
│       ├── components/           # Vue 组件
│       │   ├── MainLayout.vue    # 顶层布局（807 行，已拆 composables）
│       │   ├── editor/CodeEditor.vue   # CodeMirror 6 封装
│       │   ├── viewer/           # HexViewer / ImageViewer / ImageEditor / LogViewer
│       │   └── ...               # FindWin / DiffView / FormatConverter 等浮层
│       ├── composables/          # 业务逻辑抽离
│       │   ├── useCommands.ts    # 命令分发（150+ 命令）
│       │   ├── useFileOps.ts     # 文件 I/O + 会话 + 工作空间
│       │   └── useTailWatcher.ts # tail -f 监听（修复 EventsOn 泄漏）
│       ├── stores/               # Pinia 状态
│       │   ├── editorStore.ts    # 标签 / 书签 / 位置历史 / 宏
│       │   ├── fileStore.ts      # 当前目录 / 文件树
│       │   ├── settingStore.ts   # 配置 + 防抖持久化
│       │   └── converterTabStore.ts  # FormatConverter 内部 tab 路由
│       ├── types/index.ts        # 全局 TypeScript 类型
│       └── utils/index.ts        # 工具函数
├── .github/workflows/ci.yml      # GitHub Actions（backend + frontend 双 pipeline）
└── build/                        # 构建产物
```

## 技术栈

| 层级 | 技术 |
|------|------|
| 桌面框架 | Wails v2.12（Go 后端 + 系统 WebView2） |
| 前端 | Vue 3 + TypeScript + Composition API + `<script setup>` |
| 状态管理 | Pinia（4 个 store：editor / file / setting / converterTab） |
| UI 库 | Element Plus 2.14（zh-CN locale）+ Tailwind CSS + Lucide icons |
| 编辑器 | CodeMirror 6（13 个语言包 + autocomplete + lint + search） |
| Markdown | markdown-it + Mermaid + KaTeX + highlight.js |
| 后端语言 | Go 1.24 |
| 数据库 | SQLite via glebarez/sqlite（纯 Go，无 CGO）+ GORM |
| 编码 | saintfish/chardet + golang.org/x/text（37 种编码） |
| 脚本 | gopher-lua（带 5s 超时保护 + goroutine 独占 Close） |
| 文件监控 | fsnotify |
| CI | GitHub Actions（golangci-lint + vue-tsc + ESLint） |

## 后端分层与约定

```
Wails bound method  ←  api.Handler（薄，参数校验 + 委托）
        ↓
domain service      ←  backend/file/* 或 backend/tools/*（业务逻辑）
        ↓
pure utility        ←  backend/utils/errors.go（AppError + WrapError）
```

**核心约束：**

1. **Wails binding 解耦**：所有导出方法在 `api.Handler` 上，`App` 嵌入 `*api.Handler` 自动 promote。**禁止在 `app.go` 或 `main.go` 放业务逻辑**。
2. **错误约定**：`utils.AppError` 携带数字 code + 中文 message；**1000s 文件、2000s JSON、3000s 编码、4000s diff、5000s 配置/db**。优先用 `utils.WrapError(code, msg, cause)` 包装保留 `errors.Is/As` 链。
3. **Startup invariant**：DB 初始化失败 → fail-fast panic；DB 依赖服务（draft/recent/snippet/bookmark）保证非 nil，**API 层禁止 nil 守卫**。
4. **Goroutine 资源独占**：`defer L.Close()` 必须独占于执行 goroutine 内；goroutine 内 panic 用 `defer recover()` 兜底；`wg.Add` 先于 `go` 语句。
5. **`[]byte` 序列化坑点**：Wails 把 Go `[]byte` 序列化为 JS `[]number`，binary API 需元素级转换（见 `api_files.go`）。

## 测试

**52 个测试用例**（51 通过 + 1 个 Windows 专属 skip），覆盖 11 个包：

| 包 | 覆盖 |
|---|---|
| `utils` | AppError 错误链 `Unwrap()` / `WrapError` 包装 |
| `tools` | gopher-lua 执行/超时/并发、findreplace 查找/替换、recent 去重与限制 |
| `file` | 文件分片读取边界 |
| `config` | 配置合并 / 默认值 / v2→v3 迁移 |
| `api` | Handler 启动 fail-fast 不变量 |
| `internal/fileassoc` | 文件关联 ProgID / 扩展名清单 |
| `internal/tray` | 托盘生命周期幂等 / 线程安全 |
| `internal/concurrency` | 并发执行器边界 |

```bash
cd backend && go test ./... -count=1 -timeout 60s
```

修复 bug 时**必须先写复现测试**。

## 文档

- [产品需求文档](./EasyText—产品需求文档.md) — 产品定位、功能需求、非功能指标
- [设计文档](./EasyText—设计文档.MD) — 技术架构、API 契约、关键实现决策
- [CLAUDE.md](./CLAUDE.md) — 给 Claude Code 的项目指南（命令、架构、约定）

## 许可证

MIT License