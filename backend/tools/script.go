package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"easy-text/backend/utils"

	lua "github.com/yuin/gopher-lua"
)

// ScriptInfo 脚本信息
type ScriptInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Language    string `json:"language"` // "lua"
	Code        string `json:"code"`
	Enabled     bool   `json:"enabled"`
	MenuGroup   string `json:"menuGroup"`
	CreatedAt   string `json:"createdAt"`
}

// ScriptContext 脚本执行上下文
type ScriptContext struct {
	FilePath   string `json:"filePath"`
	Content    string `json:"content"`
	Selection  string `json:"selection"`
	CursorLine int    `json:"cursorLine"`
	CursorCol  int    `json:"cursorCol"`
	Language   string `json:"language"`
}

// ScriptResult 脚本执行结果
type ScriptResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// scriptMeta 脚本元数据（存储于 metas.json）
type scriptMeta struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	MenuGroup   string `json:"menuGroup"`
	CreatedAt   string `json:"createdAt"`
}

// ScriptService 脚本管理服务
type ScriptService struct {
	scriptsDir string
	mu         sync.RWMutex
}

// NewScriptService 创建脚本服务
func NewScriptService(scriptsDir string) *ScriptService {
	return &ScriptService{scriptsDir: scriptsDir}
}

// metaPath 返回元数据文件路径
func (s *ScriptService) metaPath() string {
	return filepath.Join(s.scriptsDir, "metas.json")
}

// loadMetas 加载所有脚本元数据
func (s *ScriptService) loadMetas() (map[string]scriptMeta, error) {
	data, err := os.ReadFile(s.metaPath())
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]scriptMeta), nil
		}
		return nil, err
	}
	var metas map[string]scriptMeta
	if err := json.Unmarshal(data, &metas); err != nil {
		return make(map[string]scriptMeta), nil
	}
	if metas == nil {
		metas = make(map[string]scriptMeta)
	}
	return metas, nil
}

// saveMetas 保存所有脚本元数据
func (s *ScriptService) saveMetas(metas map[string]scriptMeta) error {
	if err := os.MkdirAll(s.scriptsDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(metas, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.metaPath(), data, 0644)
}

// List 列出所有脚本
func (s *ScriptService) List() ([]ScriptInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.scriptsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ScriptInfo{}, nil
		}
		return nil, utils.WrapError(5001, "读取脚本目录失败", err)
	}

	metas, _ := s.loadMetas()
	results := make([]ScriptInfo, 0, len(entries))

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(name)
		if ext != ".lua" {
			continue
		}
		id := strings.TrimSuffix(name, ext)
		info, err := s.getUnlocked(id, metas)
		if err == nil {
			results = append(results, *info)
		}
	}
	return results, nil
}

// Get 获取单个脚本
func (s *ScriptService) Get(id string) (*ScriptInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metas, _ := s.loadMetas()
	return s.getUnlocked(id, metas)
}

// getUnlocked 内部获取（不加锁，需调用方持锁）
func (s *ScriptService) getUnlocked(id string, metas map[string]scriptMeta) (*ScriptInfo, error) {
	// 读取 .lua 脚本
	luaPath := filepath.Join(s.scriptsDir, id+".lua")
	if data, err := os.ReadFile(luaPath); err == nil {
		info := &ScriptInfo{
			ID:       id,
			Language: "lua",
			Code:     string(data),
		}
		s.applyMeta(info, metas)
		return info, nil
	}

	return nil, utils.NewAppError(5001, "脚本不存在", id)
}

// applyMeta 将元数据字段应用到 ScriptInfo
func (s *ScriptService) applyMeta(info *ScriptInfo, metas map[string]scriptMeta) {
	meta, ok := metas[info.ID]
	if !ok {
		// 未找到元数据，从代码首行注释提取名称
		info.Name = info.ID
		info.Enabled = true
		info.CreatedAt = time.Now().Format(time.RFC3339)

		lines := strings.SplitN(info.Code, "\n", 2)
		if len(lines) > 0 {
			trimmed := strings.TrimSpace(lines[0])
			if info.Language == "javascript" && strings.HasPrefix(trimmed, "//") {
				if name := strings.TrimSpace(strings.TrimPrefix(trimmed, "//")); name != "" {
					info.Name = name
				}
			} else if info.Language == "lua" && strings.HasPrefix(trimmed, "--") {
				if name := strings.TrimSpace(strings.TrimPrefix(trimmed, "--")); name != "" {
					info.Name = name
				}
			}
		}
		return
	}

	if meta.Name != "" {
		info.Name = meta.Name
	} else {
		info.Name = info.ID
	}
	info.Description = meta.Description
	info.Enabled = meta.Enabled
	info.MenuGroup = meta.MenuGroup
	info.CreatedAt = meta.CreatedAt
}

// Save 保存脚本（代码 + 元数据）
func (s *ScriptService) Save(script ScriptInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ext := ".lua"

	// 保存代码文件
	codePath := filepath.Join(s.scriptsDir, script.ID+ext)
	if err := os.MkdirAll(s.scriptsDir, 0755); err != nil {
		return utils.WrapError(5001, "创建脚本目录失败", err)
	}
	if err := os.WriteFile(codePath, []byte(script.Code), 0644); err != nil {
		return utils.WrapError(5001, "保存脚本失败", err)
	}

	// 保存元数据
	metas, err := s.loadMetas()
	if err != nil {
		return utils.WrapError(5001, "读取元数据失败", err)
	}

	createdAt := script.CreatedAt
	if createdAt == "" {
		// 保留已有的创建时间
		if existing, ok := metas[script.ID]; ok && existing.CreatedAt != "" {
			createdAt = existing.CreatedAt
		} else {
			createdAt = time.Now().Format(time.RFC3339)
		}
	}

	metas[script.ID] = scriptMeta{
		Name:        script.Name,
		Description: script.Description,
		Enabled:     script.Enabled,
		MenuGroup:   script.MenuGroup,
		CreatedAt:   createdAt,
	}

	return s.saveMetas(metas)
}

// Delete 删除脚本（代码 + 元数据）
func (s *ScriptService) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 删除代码文件
	os.Remove(filepath.Join(s.scriptsDir, id+".js"))
	os.Remove(filepath.Join(s.scriptsDir, id+".lua"))

	// 删除元数据
	metas, err := s.loadMetas()
	if err == nil {
		delete(metas, id)
		_ = s.saveMetas(metas)
	}

	return nil
}

// Execute 执行脚本
func (s *ScriptService) Execute(id string, ctx ScriptContext) (*ScriptResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metas, _ := s.loadMetas()
	info, err := s.getUnlocked(id, metas)
	if err != nil {
		return nil, err
	}

	switch info.Language {
	case "lua":
		return s.executeLua(info.Code, ctx)
	default:
		return &ScriptResult{
			Success: false,
			Error:   "不支持的脚本语言: " + info.Language,
		}, nil
	}
}

// luaTimeout 单个脚本执行的最长运行时间。
// 超出后不再强制中断（gopher-lua 不支持协作式取消），
// 但保证 goroutine 与 Lua VM 生命周期一致：Close 由执行 goroutine 独占负责。
const luaTimeout = 5 * time.Second

// executeLua 通过 gopher-lua 执行 Lua 脚本。
//
// 并发与资源管理：
//   - Lua VM 由执行 goroutine 独占创建并负责 Close，杜绝外层 + goroutine 双重 Close。
//   - 超时分支不再调用 L.Close()，避免对正在执行的 VM 二次 Close 触发 panic。
//   - 超时后 goroutine 仍会运行直到 L.DoString 自然返回（gopher-lua 已知限制），
//     期间 result.Output 的写入对已返回给前端的结果无副作用。
func (s *ScriptService) executeLua(code string, ctx ScriptContext) (*ScriptResult, error) {
	L := lua.NewState()
	result := &ScriptResult{Success: true}

	// 设置 editor 表
	editorTbl := L.NewTable()
	L.SetField(editorTbl, "getContent", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(ctx.Content))
		return 1
	}))
	L.SetField(editorTbl, "setContent", L.NewFunction(func(L *lua.LState) int {
		text := L.CheckString(1)
		ctx.Content = text
		result.Output = text
		return 0
	}))
	L.SetField(editorTbl, "getSelection", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(ctx.Selection))
		return 1
	}))
	L.SetField(editorTbl, "replaceSelection", L.NewFunction(func(L *lua.LState) int {
		result.Output = L.CheckString(1)
		return 0
	}))
	L.SetField(editorTbl, "insertText", L.NewFunction(func(L *lua.LState) int {
		result.Output = L.CheckString(1)
		return 0
	}))
	L.SetField(editorTbl, "getCursorLine", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(ctx.CursorLine))
		return 1
	}))
	L.SetField(editorTbl, "getCursorColumn", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(ctx.CursorCol))
		return 1
	}))
	L.SetField(editorTbl, "getFilePath", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(ctx.FilePath))
		return 1
	}))
	L.SetField(editorTbl, "getLanguage", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(ctx.Language))
		return 1
	}))
	L.SetGlobal("editor", editorTbl)

	// 设置 utils 表
	utilsTbl := L.NewTable()
	L.SetField(utilsTbl, "showMessage", L.NewFunction(func(L *lua.LState) int {
		return 0 // no-op in backend
	}))
	L.SetField(utilsTbl, "clipboardCopy", L.NewFunction(func(L *lua.LState) int {
		return 0 // no-op in backend
	}))
	L.SetGlobal("utils", utilsTbl)

	// 执行 goroutine：唯一负责 L.Close()
	done := make(chan error, 1)
	go func() {
		defer L.Close() // goroutine 内独占关闭，避免与超时分支双 Close
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("Lua 执行异常: %s", formatPanic(r))
			}
		}()
		done <- L.DoString(code)
	}()

	select {
	case err := <-done:
		if err != nil {
			result.Success = false
			result.Error = err.Error()
		}
	case <-time.After(luaTimeout):
		// 超时：仅标记失败，不强制关闭 VM。
		// goroutine 会在 L.DoString 返回后 defer Close，资源仍可释放。
		result.Success = false
		result.Error = fmt.Sprintf("Lua 执行超时（%s）", luaTimeout)
	}

	return result, nil
}

func formatPanic(r interface{}) string {
	if err, ok := r.(error); ok {
		return err.Error()
	}
	return strings.TrimSpace(jsonMarshal(r))
}

func jsonMarshal(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
