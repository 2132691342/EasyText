# EasyText

> 轻量级 Windows 桌面文本编辑器 · Wails + Vue 3 + TypeScript

**EasyText** 是一款面向开发者的 Windows 桌面编辑器。冷启动 < 1 秒、安装包 ~11MB，内置 JSON 工具 / Diff 对比 / 格式转换 / 编码转换 / Markdown 预览 / Lua 脚本 / 文件监控等高频开发工具，开箱即用。

![Windows 10/11](https://img.shields.io/badge/platform-Windows%2010%2F11-blue) ![Wails v2.12](https://img.shields.io/badge/Wails-v2.12-green) ![Go 1.24](https://img.shields.io/badge/Go-1.24-00ADD8) ![Vue 3](https://img.shields.io/badge/Vue-3.4-42b883) ![License MIT](https://img.shields.io/badge/license-MIT-yellow)

## 下载安装

安装包已随仓库发布，可直接使用：

- **安装包**：[`build/bin/EasyText-amd64-installer.exe`](build/bin/EasyText-amd64-installer.exe) — 双击安装，可选注册为常见文本类型的编辑器
- **免安装单文件**：[`build/bin/EasyText.exe`](build/bin/EasyText.exe) — 绿色运行，零依赖

> 需要 Windows 10 1809+ / 11（系统已内置 WebView2）。

## 核心特性

- **轻量原生**：系统 WebView2 渲染，单 EXE ~24MB，零依赖
- **代码编辑**：CodeMirror 6 引擎，60+ 语法高亮、代码折叠、多光标、括号匹配
- **格式转换**：JSON / YAML / TOML / XML 互转；JSONPath 查询；JSON 转结构体（Go/TS/Java/Python 等）；JSON 结构化 Diff
- **开发者工具箱**：LCS 行级 Diff + 字符级高亮、38 种编码检测与转换、批量查找/替换/重命名、MD5/SHA 哈希、列块模式（Alt+X）
- **Markdown**：编辑/分屏/预览三模式，Mermaid 图表 + KaTeX 公式 + 代码块高亮，一键导出 HTML
- **脚本系统**：内嵌 Lua 脚本管理器，5s 超时保护与并发安全
- **文件监控**：实时 tail -f 文件变化并自动重载
- **会话恢复**：启动时自动恢复上次打开的文件、工作空间、光标位置
- **外观**：19 套主题 + 暗色模式，跟随系统切换
- **侧栏面板**：代码片段、全局书签、函数列表、文件监控、日志查看
- **正则测试器**：实时匹配测试，支持捕获组提取
- **系统托盘**：关闭到托盘常驻，可从托盘恢复或彻底退出
- **文件关联**：安装器可选注册为常见文本类型的编辑器，出现在右键「打开方式」列表（不抢占默认）

## 快速开始

### 环境要求

- **Go** 1.24+ · **Node.js** 18+ · **Wails CLI**（`go install github.com/wailsapp/wails/v2/cmd/wails@latest`）
- **Windows 10 1809+ / 11**

### 开发与构建

```bash
wails dev                                   # 开发模式（Vite HMR + Go 后端）
wails build                                 # 产出 build/bin/EasyText.exe
wails build -platform windows/amd64 --nsis  # NSIS 安装包
```

### 本地验证

```bash
make ci-build            # backend build/vet/test + frontend typecheck/build
```

| 命令 | 作用 |
|---|---|
| `make backend` | 后端 `go build` + `go vet` |
| `make frontend` | 前端安装依赖 + `typecheck` + `build` |
| `make test` | 后端 `go test ./...` |
| `make lint` | 前端 ESLint |
| `make ci-build` | CI 等价检查 |
| `make clean` | 清理构建缓存 |

## 常用快捷键

| 快捷键 | 功能 | 快捷键 | 功能 |
|--------|------|--------|------|
| `Ctrl+O` / `Ctrl+S` | 打开 / 保存 | `Ctrl+Z` / `Ctrl+Y` | 撤销 / 重做 |
| `Ctrl+F` / `Ctrl+H` | 查找 / 替换 | `Ctrl+G` | 转到行 |
| `F3` / `Shift+F3` | 查找下一个 / 上一个 | `Ctrl+/` | 行注释 |
| `Ctrl+D` / `Ctrl+L` | 复制行 / 删除行 | `Alt+X` | 列编辑模式 |
| `Ctrl+Shift+F` | 全局搜索 | `Ctrl+Shift+D` | 目录查找 |
| `F2` | 切换书签 | `F11` / `Ctrl+P` | 全屏 / 打印 |

## 项目结构

```
EasyText/
├── main.go / app.go              # Wails 入口（App 嵌入 *api.Handler）
├── backend/
│   ├── api/                      # Wails 绑定层（薄，按特性分文件）
│   ├── tools/                    # 业务服务（json/diff/encoding/findreplace/…）
│   ├── file/                     # 文件域（reader/writer/tree/watcher）
│   ├── config/                   # AppConfig + SQLite
│   └── …                         # tray / closepolicy / singleinstance / fileassoc
├── frontend/src/
│   ├── components/               # Chrome 五层 + ModalOverlay + 各功能浮窗
│   │   └── editor/               # CodeEditor 容器 + composables + ext
│   ├── composables/              # useCommands / useFileOps / useTailWatcher
│   ├── stores/                   # Pinia：editor / file / setting / converterTab
│   └── style.css                 # 设计令牌 --et-*
├── docs/ARCHITECTURE.md          # 架构与实现说明（读代码前先看这个）
└── .github/workflows/ci.yml      # CI：backend(windows) + frontend
```

## 技术栈

| 层级 | 技术 |
|------|------|
| 桌面框架 | Wails v2.12（Go 后端 + 系统 WebView2） |
| 前端 | Vue 3 + TypeScript + Composition API + `<script setup>` |
| 状态管理 | Pinia |
| UI 库 | Element Plus（zh-CN）+ Tailwind CSS + Lucide icons |
| 编辑器 | CodeMirror 6 |
| Markdown | markdown-it + Mermaid + KaTeX |
| 后端语言 | Go 1.24 |
| 数据库 | SQLite via glebarez/sqlite（纯 Go）+ GORM |
| 编码 | chardet + golang.org/x/text（38 种） |
| 脚本 | gopher-lua（5s 超时保护） |
| 文件监控 | fsnotify |

## 测试

后端 23 个测试，聚焦复杂流程与长链路（查找替换选项矩阵与目录级管线、编码往返与 BOM、脚本超时与并发、各服务生命周期、配置迁移、启动装配），全量运行 < 5 秒：

```bash
cd backend && go test ./... -count=1 -timeout 120s
```

## 文档

- [架构与实现](./docs/ARCHITECTURE.md) — 整体架构、模块设计、关键实现
- [CLAUDE.md](./CLAUDE.md) — 开发与协作速查（命令、约定、测试）

## 许可证

MIT License
