package api

import (
	"easy-text/backend/tools"
	"easy-text/internal/fileassoc"
)

// =============================================================================
// Handler 拆分骨架（ch11 ISP + 后续渐进提取）
// =============================================================================
//
// 历史：原 handler.go 单体持有 14 个 service 字段、73+ 公开方法，
// 任何 panic 都会拖垮整个 Wails App。单体也违反 ISP：调用方拿到的接口太大，
// 单元测试无法用 stub 替换其中一个 service。
//
// 本文件定义"小组合器"模式作为拆分骨架。每个子 handler：
//   - 仅持有自己需要的 service 依赖
//   - 暴露领域相关方法（通过 Go 方法提升对 Wails 自动可见）
//   - 单独可测（构造器接受接口而非具体 *tools.RecentService）
//
// 当前已落地的子模块：
//   - RecentHandler   : 最近文件 / 文件夹（5 个方法）—— 已迁移
//   - FileAssocHandler: 运行时文件关联（3 个方法）—— 已迁移
//   - SearchHandler   : 查找替换 + 文档对比（11 个方法）—— 已迁移
// 未来拆分候选（每个按本文件模板即可独立完成）：
//   - FileHandler     : OpenFile / SaveFile / ReadFileBytes / Convert
//   - WorkspaceHandler: Draft / Snippet / Bookmark / Session / Workspace
//   - ToolsHandler    : JSON / Encoding / Convert / Hash / Regex（Diff 已并入 Search）
//   - MacrosHandler   : 宏录制 / 回放
//
// 拆分原则：
//   - 一个子 handler ≤ 200 行（不含表数据）
//   - 公共方法数 ≤ 20
//   - 子 handler 之间的依赖通过接口（不是直接指针）
//
// =============================================================================

// RecentHandler 最近文件 / 文件夹子系统。
//
// 拆分理由：原 Handler.GetRecentFiles 暴露给前端"打开方式"菜单直接调用；
// 隔离为子模块后能用 stub 替换 recentService 写单测，验证 Add 与最近
// 列表的滚动策略。详细测试见 backend/tools/recent_test.go。
type RecentHandler struct {
	recentService *tools.RecentService
}

// NewRecentHandler 构造最近文件子 handler。recentService 由 Startup 注入，
// 不直接持有 *tools.RecentService 指针便于未来切接口。
func NewRecentHandler(svc *tools.RecentService) *RecentHandler {
	return &RecentHandler{recentService: svc}
}

// GetRecentFiles 获取最近打开的文件列表。
func (h *RecentHandler) GetRecentFiles() ([]tools.RecentEntryResult, error) {
	return h.recentService.GetFiles()
}

// GetRecentFolders 获取最近打开的文件夹列表。
func (h *RecentHandler) GetRecentFolders() ([]tools.RecentEntryResult, error) {
	return h.recentService.GetFolders()
}

// AddRecentEntry 添加最近访问记录。注入 false 表示文件，true 表示文件夹。
func (h *RecentHandler) AddRecentEntry(path string, isFolder bool) error {
	return h.recentService.Add(path, isFolder)
}

// ClearRecentFiles 清除所有最近文件记录。
func (h *RecentHandler) ClearRecentFiles() error {
	return h.recentService.ClearFiles()
}

// ClearRecentFolders 清除所有最近文件夹记录。
func (h *RecentHandler) ClearRecentFolders() error {
	return h.recentService.ClearFolders()
}

// FileAssocHandler 运行时文件关联子系统（便携模式支持）。
//
// 拆分理由：依赖 registry（仅 Windows 平台）。隔离到子模块后，
// 跨平台编译时可以用 build tag 隔离，不必在主 Handler 里塞 ifdef。
type FileAssocHandler struct{}

// NewFileAssocHandler 构造文件关联子 handler。无运行时依赖（全部为包级函数）。
func NewFileAssocHandler() *FileAssocHandler {
	return &FileAssocHandler{}
}

// RegisterFileAssoc 把 EasyText 注册为当前用户的文本文件关联。
// 仅写入 HKCU（不需要管理员权限）。
func (h *FileAssocHandler) RegisterFileAssoc() ([]string, error) {
	return fileassoc.Register(fileassoc.DefaultExtensions)
}

// UnregisterFileAssoc 反向清理 RegisterFileAssoc 写入的键值。
func (h *FileAssocHandler) UnregisterFileAssoc() error {
	return fileassoc.Unregister(fileassoc.DefaultExtensions)
}

// IsFileAssocRegistered 查询当前是否已运行时注册（Setting 页开关状态）。
func (h *FileAssocHandler) IsFileAssocRegistered() bool {
	return fileassoc.IsRegistered()
}

// SearchHandler 查找替换 + 文档对比子系统。
//
// 拆分理由：findService / diffTool 是两个独立领域，原版挂在 Handler 上
// 让 diff 与 find 混在 14 个 service 字段里。隔离后：
//   - SearchHandler 只持有 findService + diffTool，方法集中在查找替换 + 对比；
//   - 后续 find 换成流式实现 / diff 换算法，只需要动这个子 handler，
//     不惊动 Handler 生命周期。
type SearchHandler struct {
	findService *tools.FindReplaceService
	diffTool    *tools.DiffTool
}

// NewSearchHandler 构造查找替换 + 对比子 handler。
func NewSearchHandler(fs *tools.FindReplaceService, dt *tools.DiffTool) *SearchHandler {
	return &SearchHandler{findService: fs, diffTool: dt}
}
