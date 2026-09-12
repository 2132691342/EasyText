# Frontend

EasyText 前端：Vue 3 + TypeScript + Vite，运行在 Wails 的 WebView2 容器内。

## 技术栈

| 类别 | 库 |
|------|-----|
| 框架 | Vue 3（Composition API + `<script setup>`） |
| 语言 | TypeScript |
| 构建 | Vite |
| 状态管理 | Pinia |
| UI 组件 | Element Plus（zh-CN 本地化） |
| 样式 | Tailwind CSS + 全局设计令牌 `--et-*` |
| 图标 | Lucide Vue Next |
| 代码编辑 | CodeMirror 6 |
| 预览 | pdfjs / docx-preview / xlsx / jszip / markdown-it |

## 常用命令

```bash
npm install          # 安装依赖
npm run dev          # 开发模式（配合 wails dev 的后端）
npm run build        # vue-tsc 类型检查 + Vite 生产构建
npm run lint         # ESLint
```

## 目录结构

```
src/
├── App.vue                     # 根组件：加载配置 → 挂载 MainLayout
├── main.ts                     # 入口：创建 App + 注册 Element Plus
├── style.css                   # 设计令牌 --et-* 与 .et-chrome-* 工具类
│
├── components/
│   ├── MainLayout.vue          # 顶层装配 + 全局事件接线（file:change 等）
│   ├── NddMenuBar.vue          # 菜单栏（28px）
│   ├── NddToolbar.vue          # 工具栏（30px，即时 Popover）
│   ├── TabBar.vue              # 标签栏（30px）
│   ├── StatusBar.vue           # 状态栏（24px）
│   ├── Sidebar.vue             # 侧边栏容器（文件树/面板切换）
│   ├── ModalOverlay.vue        # 所有浮窗统一容器（Teleport/遮罩/focus-trap/Esc）
│   ├── EditorArea.vue          # 编辑区（按 viewType 分发到编辑器或查看器）
│   ├── editor/
│   │   ├── CodeEditor.vue      # CodeMirror 容器：compartment 切换、就地换绑
│   │   ├── Minimap.vue         # 小地图（视口矩形 + rAF 滚动同步）
│   │   ├── composables/        # 主题/语言/补全/书签/列编辑/宏/MD 扩展模块
│   │   └── ext/                # keymap 装配 + 右键菜单数据
│   └── viewer/                 # Hex/图片/日志/取色器查看器
│   └── *.vue                   # 功能浮窗：查找替换/Diff/批量操作/设置/正则测试…
│
├── composables/
│   ├── useCommands.ts          # 命令中枢：快捷键/菜单/工具栏统一分发
│   ├── useFileOps.ts           # 打开/保存/拖放等文件操作
│   └── useTailWatcher.ts       # 文件尾随监控
│
├── stores/                     # Pinia
│   ├── editorStore.ts          # 标签页生命周期、脏标记、宏状态
│   ├── fileStore.ts            # 目录树状态
│   ├── settingStore.ts         # 配置镜像（主题/字号/…）
│   └── converterTabStore.ts    # 格式转换器页签
│
├── types/index.ts              # FileInfo / EditorTab / TabViewType 等全局类型
└── wailsjs/                    # Wails 绑定（gen-bindings.cjs 生成，已提交）
```

## 关键约定

- **样式令牌**：颜色/间距/字号只用 `--et-*` 变量，禁止新增字面色。
- **浮窗**：一律包 `<ModalOverlay>`，组件只写内容，不自写 Teleport/遮罩/Esc。
- **命令**：新功能挂到 `useCommands` 的命令表，快捷键经 `ext/ext-keymap.ts`
  的 `CMD_ALIASES` 映射，保证菜单与按键同路径。
- **二进制读写**：统一走 `ReadFileBytes` → `normalizeBytes()` → `Uint8Array`；
  保存用 `SaveFileBytes(path, data: number[])`（Wails 把 Go `[]byte` 序列化为 `number[]`）。
- **图标**：统一使用 `lucide-vue-next` 的 PascalCase 名称。
- **类型**：所有 `<script setup>` 显式标注 `lang="ts"`。

整体架构见 [`../docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md)。
