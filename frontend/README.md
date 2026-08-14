# Frontend 目录

EasyText 的前端代码，基于 **Vue 3 + TypeScript + Vite**。

## 技术栈

| 类别 | 库 |
|------|-----|
| 框架 | Vue 3.4（Composition API + `<script setup>`） |
| 语言 | TypeScript 5.3 |
| 构建 | Vite 5 |
| 状态管理 | Pinia 2 |
| UI 组件 | Element Plus 2.14（zh-CN 本地化） |
| 样式 | Tailwind CSS 3 |
| 图标 | Lucide Vue Next |
| 代码编辑 | CodeMirror 6（basicSetup + 多语言扩展） |
| PDF 渲染 | pdfjs-dist |
| Word 渲染 | docx-preview |
| Excel 解析/编辑 | xlsx (SheetJS) |
| PPT 解析 | jszip |
| Markdown | markdown-it |

## 常用命令

```bash
# 安装依赖
npm install

# 开发模式（需先启动 Wails dev）
npm run dev

# 类型检查 + 生产构建
npm run build

# 本地预览构建产物
npm run preview
```

## 目录结构

```
src/
├── App.vue                     # 根组件：加载配置 → 挂载 MainLayout
├── main.ts                     # 入口：创建 App + 注册 Element Plus
├── style.css                   # 全局样式
│
├── components/                 # 业务组件
│   ├── MainLayout.vue          # 主布局
│   ├── Toolbar.vue             # 工具栏（Element Plus 下拉菜单）
│   ├── Sidebar.vue             # 侧边栏
│   ├── FileTree.vue            # 递归文件树
│   ├── EditorArea.vue          # 编辑区（按 viewType 分发）
│   ├── TabBar.vue              # 标签栏（右键菜单 + 行内重命名）
│   ├── StatusBar.vue           # 状态栏
│   ├── WelcomeScreen.vue       # 欢迎页
│   ├── FindReplace.vue         # 查找替换
│   ├── SettingsView.vue        # 设置面板（Element Plus 表单）
│   ├── DiffView.vue            # 文档对比
│   ├── FormatConverter.vue     # 格式转换器
│   ├── editor/
│   │   └── CodeEditor.vue      # CodeMirror 6 编辑器
│   └── viewer/                 # 文件查看器
│       ├── PdfViewer.vue
│       ├── ExcelViewer.vue
│       ├── WordViewer.vue
│       ├── PptViewer.vue
│       └── ImageViewer.vue
│
├── stores/                     # Pinia stores
│   ├── editorStore.ts          # 标签页生命周期
│   ├── fileStore.ts            # 文件/目录状态
│   └── settingStore.ts         # 设置/主题状态
│
├── types/                      # TypeScript 类型定义
│   └── index.ts                # FileInfo / EditorTab / TabViewType / ...
│
├── utils/                      # 工具函数
│   └── index.ts                # normalizeBytes / getTabViewType / ...
│
└── wailsjs/                    # Wails 自动生成的 TS 绑定（勿手动修改）
    └── go/main/App.{js,d.ts}
```

## 关键约定

- **二进制文件读取**：统一走 `ReadFileBytes` → `normalizeBytes()` → `Uint8Array`
- **二进制文件保存**：通过 `SaveFileBytes(path, data: number[])`（Wails 把 Go `[]byte` 序列化为 `[]number`）
- **浮层**：`SettingsView`/`DiffView`/`FormatConverter` 均用 `<Teleport to="body">` 渲染到根，避免父容器 `overflow` 截断
- **图标**：统一使用 `lucide-vue-next` 的 PascalCase 名称（如 `FileText` / `Folder` / `Trash2`）
- **类型注解**：所有 `<script setup>` 必须显式标注 `lang="ts"`

详细技术设计见 [`../EasyText—设计文档.MD`](../EasyText—设计文档.MD) 第 4 节。
