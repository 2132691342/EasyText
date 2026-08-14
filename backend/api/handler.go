package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"easy-text/backend/config"
	"easy-text/backend/file"
	"easy-text/backend/tools"
	"easy-text/backend/utils"
)

// Handler 处理前后端 API 交互，是 Wails 绑定的核心结构体。
// 仿 notepad-- ccnotepad.h 的架构设计，按功能模块组织：
//
//	api_files.go      - 文件 I/O、Office 转换
//	api_dirs.go       - 目录操作
//	api_config.go     - 配置 & 系统设置
//	api_json.go       - JSON 工具
//	api_diff.go       - 文档对比
//	api_encoding.go   - 编码工具
//	api_convert.go    - 格式转换
//	api_dialogs.go    - 对话框 & 系统信息
//	api_macro.go      - 宏录制/回放
//	api_find.go       - 查找替换服务
//	api_draft.go      - 草稿系统（🆕 V2.0.0）
//	api_snippet.go    - 代码片段（🆕 V2.0.0）
//	api_bookmark.go   - 书签持久化（🆕 V2.0.0）
//
// 渐进拆分（ch11 ISP）：
//
//	Handler 通过嵌入 *RecentHandler / *FileAssocHandler 等"小组合器"
//	把领域方法提升为自身可被 Wails 反射的成员。详细拆分骨架见
//	handler_subsystem.go，包含未来可继续拆分的所有候选子模块。
type Handler struct {
	Ctx context.Context

	// 已拆分的子 handler：其方法通过 Go 嵌入提升被 Wails 反射看到。
	*RecentHandler
	*FileAssocHandler
	*SearchHandler

	// 仍由当前 Handler 直接持有 service；后续按子模块继续拆分。
	fileReader     *file.FileReader
	fileWriter     *file.FileWriter
	treeBuilder    *file.TreeBuilder
	fileWatcher    *file.FileWatcher
	jsonTool       *tools.JSONTool
	encodingTool   *tools.EncodingTool
	convertTool    *tools.ConvertTool
	macroService   *tools.MacroService
	compareService *tools.CompareService

	// 🆕 V2.0.0 仍直接持有的服务（待后续拆分到 Draft/Snippet/Bookmark/Workspace 子 handler）：
	draftService    *tools.DraftService
	snippetService  *tools.SnippetService
	bookmarkService *tools.BookmarkService
	scriptService   *tools.ScriptService
}

// NewHandler 创建 Handler 实例，初始化所有领域服务。
//
// 拆分过程中保留核心生命周期：DB 依赖服务在 Startup 阶段 fail-fast 保证非 nil，
// 子 handler 通过指针接入（不复制 service）。
func NewHandler() *Handler {
	fw, err := file.NewFileWatcher()
	if err != nil {
		panic(fmt.Sprintf("Failed to create file watcher (required for tail-f / file change events): %v", err))
	}
	return &Handler{
		fileReader:       file.NewFileReader(100 * 1024 * 1024), // 100MB max
		fileWriter:       file.NewFileWriter(),
		treeBuilder:      file.NewTreeBuilder(),
		fileWatcher:      fw,
		jsonTool:         tools.NewJSONTool(),
		encodingTool:     tools.NewEncodingTool(),
		convertTool:      tools.NewConvertTool(),
		macroService:     tools.NewMacroService(),
		compareService:   tools.NewCompareService(),
		FileAssocHandler: NewFileAssocHandler(),
		SearchHandler:    NewSearchHandler(tools.NewFindReplaceService(), tools.NewDiffTool()),
		// RecentHandler 在 Startup 阶段依赖 DB 注入，此处先保留 nil；
		// Startup 中通过 h.RecentHandler = NewRecentHandler(recSvc) 注入。
	}
}

// Startup 应用启动时调用，初始化日志、数据库和配置。
//
// 失败模式：
//   - 数据库初始化失败：fail-fast panic。DB 是草稿/书签/最近访问/片段的依赖，
//     任一核心服务不可用时启动半残应用反而误导用户，应在启动期就失败。
//   - 配置/日志/脚本目录失败：仅记录日志，应用以可用子集继续运行。
//   - scriptService 不依赖 DB，始终构造。
//
// 通过 fail-fast，依赖 DB 的 4 个服务在 Startup 之后保证非 nil，
// 各 API 方法无需重复 nil 守卫。
//
// 🆕 配置依赖注入（Step 5）：构造 config.Source 适配 *ConfigManager，
// 把全局单例通过接口注入到 RecentService，便于未来切接口做单测。
func (h *Handler) Startup(ctx context.Context) {
	h.Ctx = ctx

	// 初始化日志
	if err := utils.InitLogger(""); err != nil {
		fmt.Println("Failed to initialize logger:", err)
	}

	// 初始化配置目录
	configDir, err := config.GetConfigDir()
	if err != nil {
		utils.Log.Error("Failed to get config directory: %v", err)
	}

	if configDir != "" {
		// 初始化数据库（fail-fast）
		dbPath := filepath.Join(configDir, "easytext.db")
		if err := config.InitDatabase(dbPath); err != nil {
			panic(fmt.Sprintf("Failed to initialize database (required for v2.0.0 services): %v", err))
		}

		// 初始化配置（非致命）
		configPath := filepath.Join(configDir, "config.json")
		if err := config.InitConfig(configPath); err != nil {
			utils.Log.Error("Failed to initialize config: %v", err)
		}

		// 通过 Source 接口注入配置；不再让业务服务直读全局。
		cfgSource := config.NewSource(config.Config)

		// 初始化依赖 DB 的服务（config.DB 已由 InitDatabase 保证非 nil）
		recSvc := tools.NewRecentService(config.DB, cfgSource)
		h.RecentHandler = NewRecentHandler(recSvc)
		h.draftService = tools.NewDraftService(config.DB)
		h.snippetService = tools.NewSnippetService(config.DB)
		h.bookmarkService = tools.NewBookmarkService(config.DB)

		// 确保脚本目录存在
		scriptsDir := filepath.Join(configDir, "scripts")
		if err := os.MkdirAll(scriptsDir, 0755); err != nil {
			utils.Log.Error("Failed to create scripts directory: %v", err)
		}
		// scriptService 不依赖 DB，始终构造
		h.scriptService = tools.NewScriptService(scriptsDir)
	} else {
		// 配置目录不可用时，scriptService 仍以临时目录兜底（最小可用集）
		h.scriptService = tools.NewScriptService(filepath.Join(os.TempDir(), "easytext-scripts"))
	}

	utils.Log.Info("Application started successfully (v2.0.0)")
}

// Shutdown 应用关闭时调用
func (h *Handler) Shutdown(ctx context.Context) {
	utils.Log.Info("Application shutting down")
}
