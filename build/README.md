# Build 目录

存放 EasyText 的所有构建产物和平台资源。

## 目录结构

```
build/
├── bin/                          # 编译输出
│   ├── easy-text.exe             # 主可执行文件（~24MB）
│   └── EasyText-amd64-installer.exe  # NSIS 安装包（~11MB）
├── windows/                      # Windows 平台资源
│   ├── icon.ico                  # 应用图标
│   ├── info.json                 # 详细信息（右键属性 → 详细信息）
│   ├── wails.exe.manifest        # 应用清单
│   └── installer/                # NSIS 安装包模板
│       ├── project.nsi           # NSIS 脚本
│       └── wails_tools.nsh       # NSIS 宏定义
└── darwin/                       # macOS 平台资源（占位）
    ├── Info.plist
    └── Info.dev.plist
```

## Windows 构建

```bash
# 仅生成可执行文件
wails build -platform windows/amd64

# 生成 NSIS 安装包（需先安装 NSIS）
winget install NSIS.NSIS
wails build -platform windows/amd64 --nsis
```

## NSIS 关键配置（`installer/project.nsi`）

| 配置项 | 值 |
|--------|-----|
| 安装目录 | `$PROGRAMFILES64\EasyText\EasyText` |
| 输出文件名 | `EasyText-${ARCH}-installer.exe` |
| WebView2 | 自动引导安装 |
| 桌面 / 开始菜单快捷方式 | 自动创建 |
| 卸载 | 清理程序目录、WebView2 数据、快捷方式 |

详细构建说明见 [`../EasyText—设计文档.MD` 第 7 节](../EasyText—设计文档.MD)。
