# Build

存放 EasyText 的构建产物与 Windows 平台资源。

```
build/
├── bin/                          # 编译输出
│   ├── EasyText.exe              # 主程序（~24MB，免安装）
│   └── EasyText-amd64-installer.exe  # NSIS 安装包（~11MB）
└── windows/                      # Windows 平台资源
    ├── icon.ico                  # 应用图标
    ├── info.json                 # 版本信息（右键属性）
    ├── wails.exe.manifest        # 应用清单
    └── installer/                # NSIS 安装包模板
        ├── project.nsi           # NSIS 脚本
        └── wails_tools.nsh       # NSIS 宏定义
```

## 构建命令

```bash
wails build -platform windows/amd64           # 仅生成 exe
wails build -platform windows/amd64 --nsis    # 生成 NSIS 安装包（需安装 NSIS：winget install NSIS.NSIS）
```

## NSIS 关键配置（installer/project.nsi）

| 配置项 | 值 |
|--------|-----|
| 安装目录 | `$PROGRAMFILES64\EasyText\EasyText` |
| 输出文件名 | `EasyText-${ARCH}-installer.exe` |
| WebView2 | 自动引导安装 |
| 快捷方式 | 桌面 + 开始菜单，可选文件关联 |
| 卸载 | 清理程序目录、WebView2 数据、快捷方式与注册信息 |
