package utils

import "fmt"

// AppError represents a structured application error.
//
// 与标准库 error 协作：
//   - 实现 Unwrap() 返回 Cause，使 errors.Is / errors.As 能穿透到根因。
//   - Cause 不参与 JSON 序列化，避免泄露内部细节给前端。
//   - Error() 优先展示 Cause（用 %v 递归 Error()，便于嵌套包装）。
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Cause   error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// Unwrap 返回底层错误，使 errors.Is / errors.As 可穿透到根因。
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError creates a new AppError without cause.
// 仅用于静态信息错误（如"文件不存在"单例）。需要携带根因请用 WrapError。
func NewAppError(code int, message string, details string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// WrapError creates a new AppError that wraps an underlying error,
// preserving the error chain via Unwrap(). 调用方应优先使用此函数。
func WrapError(code int, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// 以下所有 ErrXxx 是包级**不可变单例**：跨函数共享同一指针，用于
// `errors.Is(err, ErrFileNotFound)` 的指针相等比较。
//
// ⚠️ 不可变契约（Go 无法强制字段只读，靠约定维护）：
//   调用方**禁止**通过 `err.(*AppError).Details = ...` 之类操作修改这些单例的
//   字段——一旦被改，全进程所有引用该单例的错误都会被污染。
//   需要携带运行时信息的错误，请改用 NewAppError / WrapError 新建实例，
//   不要复用这里的单例。
//
// 保留为导出 var 而非工厂函数的原因：大量调用点依赖 `== ErrXxx` / `errors.Is` 的
// 指针语义；改成 `func ErrFileNotFound() *AppError` 会破坏这些比较。
//
// File errors (1000-1999)
var (
	ErrFileNotFound      = &AppError{Code: 1001, Message: "文件不存在"}
	ErrPermissionDenied  = &AppError{Code: 1002, Message: "权限不足"}
	ErrFileTooLarge      = &AppError{Code: 1003, Message: "文件过大"}
	ErrFileAlreadyOpen   = &AppError{Code: 1004, Message: "文件已打开"}
	ErrInvalidPath       = &AppError{Code: 1005, Message: "无效的路径"}
	ErrDirectoryNotFound = &AppError{Code: 1006, Message: "目录不存在"}
)

// JSON errors (2000-2999)
var (
	ErrInvalidJSON      = &AppError{Code: 2001, Message: "无效的JSON格式"}
	ErrJSONParseFailed  = &AppError{Code: 2002, Message: "JSON解析失败"}
	ErrJSONFormatFailed = &AppError{Code: 2003, Message: "JSON格式化失败"}
)

// Encoding errors (3000-3999)
var (
	ErrEncodingFailed       = &AppError{Code: 3001, Message: "编码转换失败"}
	ErrUnsupportedEncoding  = &AppError{Code: 3002, Message: "不支持的编码"}
	ErrEncodingDetectFailed = &AppError{Code: 3003, Message: "编码检测失败"}
)

// Diff errors (4000-4999)
var (
	ErrDiffFailed = &AppError{Code: 4001, Message: "文档对比失败"}
)

// Config errors (5000-5999)
var (
	ErrConfigLoadFailed   = &AppError{Code: 5001, Message: "配置加载失败"}
	ErrConfigSaveFailed   = &AppError{Code: 5002, Message: "配置保存失败"}
	ErrDatabaseInitFailed = &AppError{Code: 5003, Message: "数据库初始化失败"}
)

// Compare errors (6000-6999)
var (
	ErrDirCompareFailed = &AppError{Code: 6001, Message: "目录对比失败"}
	ErrBinCompareFailed = &AppError{Code: 6002, Message: "二进制对比失败"}
)

// System/terminal errors (7000-7999)
var (
	ErrTerminalLaunchFailed = &AppError{Code: 7001, Message: "打开命令行失败"}
)

// Draft/Bookmark errors (8000-8999) 🆕 V2.0.0
var (
	ErrDraftSaveFailed    = &AppError{Code: 8001, Message: "草稿保存失败"}
	ErrDraftLoadFailed    = &AppError{Code: 8002, Message: "草稿加载失败"}
	ErrDraftConflict      = &AppError{Code: 8003, Message: "草稿冲突"}
	ErrBookmarkNotFound   = &AppError{Code: 8004, Message: "书签不存在"}
	ErrBookmarkSaveFailed = &AppError{Code: 8005, Message: "书签保存失败"}
	ErrBookmarkLoadFailed = &AppError{Code: 8006, Message: "书签加载失败"}
)

// Remote/Script/Git errors (9000-9999) 🆕 V2.0.0
var (
	ErrRemoteConnectFailed = &AppError{Code: 9001, Message: "远程连接失败"}
	ErrRemoteAuthFailed    = &AppError{Code: 9002, Message: "远程认证失败"}
	ErrRemoteDisconnected  = &AppError{Code: 9003, Message: "远程连接已断开"}
	ErrRemoteFileNotFound  = &AppError{Code: 9004, Message: "远程文件不存在"}
	ErrScriptCompileFailed = &AppError{Code: 9005, Message: "脚本编译失败"}
	ErrScriptExecuteFailed = &AppError{Code: 9006, Message: "脚本执行失败"}
	ErrScriptTimeout       = &AppError{Code: 9007, Message: "脚本执行超时"}
	ErrGitNotInstalled     = &AppError{Code: 9008, Message: "Git 未安装"}
	ErrGitOperationFailed  = &AppError{Code: 9009, Message: "Git 操作失败"}
)
