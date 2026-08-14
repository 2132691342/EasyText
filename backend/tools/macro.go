package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"easy-text/backend/utils"
)

// MacroStep 表示宏录制的一个步骤（仿 notepad-- 宏系统）
type MacroStep struct {
	Type      string `json:"type"` // insert, delete, replace, selection, cursor, command, find
	Text      string `json:"text,omitempty"`
	From      int    `json:"from,omitempty"`
	To        int    `json:"to,omitempty"`
	Anchor    int    `json:"anchor,omitempty"`
	Head      int    `json:"head,omitempty"`
	Timestamp int64  `json:"timestamp"`
	Command   string `json:"command,omitempty"`
}

// Macro 表示一个完整的宏
type Macro struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Steps      []MacroStep `json:"steps"`
	CreatedAt  int64       `json:"createdAt"`
	ModifiedAt int64       `json:"modifiedAt"`
}

// MacroService 宏录制和回放服务
type MacroService struct {
	mu           sync.RWMutex
	macros       []Macro
	storagePath  string
	isRecording  bool
	currentMacro []MacroStep
}

// NewMacroService 创建宏服务实例
func NewMacroService() *MacroService {
	configDir, err := os.UserConfigDir()
	storagePath := ""
	if err == nil {
		storagePath = filepath.Join(configDir, "easytext", "macros.json")
	}

	s := &MacroService{
		macros:      make([]Macro, 0),
		storagePath: storagePath,
	}
	s.load()
	return s
}

// StartRecording 开始录制宏
func (s *MacroService) StartRecording() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isRecording = true
	s.currentMacro = make([]MacroStep, 0)
}

// StopRecording 停止录制并保存宏
func (s *MacroService) StopRecording() *Macro {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isRecording = false
	if len(s.currentMacro) == 0 {
		return nil
	}
	now := time.Now().UnixMilli()
	macro := Macro{
		ID:         generateID(),
		Name:       fmt.Sprintf("Macro %d", len(s.macros)+1),
		Steps:      make([]MacroStep, len(s.currentMacro)),
		CreatedAt:  now,
		ModifiedAt: now,
	}
	copy(macro.Steps, s.currentMacro)
	s.macros = append(s.macros, macro)
	s.currentMacro = nil
	s.save()
	return &macro
}

// RecordStep 记录一个步骤
func (s *MacroService) RecordStep(step MacroStep) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isRecording {
		step.Timestamp = time.Now().UnixMilli()
		s.currentMacro = append(s.currentMacro, step)
	}
}

// GetMacros 获取所有保存的宏
func (s *MacroService) GetMacros() []Macro {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Macro, len(s.macros))
	copy(result, s.macros)
	return result
}

// GetMacro 获取指定宏
func (s *MacroService) GetMacro(id string) *Macro {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.macros {
		if s.macros[i].ID == id {
			m := s.macros[i]
			return &m
		}
	}
	return nil
}

// DeleteMacro 删除指定宏
func (s *MacroService) DeleteMacro(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.macros {
		if s.macros[i].ID == id {
			s.macros = append(s.macros[:i], s.macros[i+1:]...)
			s.save()
			return true
		}
	}
	return false
}

// RenameMacro 重命名宏
func (s *MacroService) RenameMacro(id, newName string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.macros {
		if s.macros[i].ID == id {
			s.macros[i].Name = newName
			s.save()
			return true
		}
	}
	return false
}

// IsRecording 是否正在录制
func (s *MacroService) IsRecording() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRecording
}

// SaveCurrentMacro 保存当前录制中的宏（用于前端直接管理宏列表）
func (s *MacroService) SaveCurrentMacro(name string) *Macro {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.currentMacro) == 0 {
		return nil
	}
	now := time.Now().UnixMilli()
	macro := Macro{
		ID:         generateID(),
		Name:       name,
		Steps:      make([]MacroStep, len(s.currentMacro)),
		CreatedAt:  now,
		ModifiedAt: now,
	}
	copy(macro.Steps, s.currentMacro)
	s.macros = append(s.macros, macro)
	s.currentMacro = nil
	s.isRecording = false
	s.save()
	return &macro
}

// save 保存宏到文件
func (s *MacroService) save() {
	if s.storagePath == "" {
		return
	}
	dir := filepath.Dir(s.storagePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		utils.Log.Error("Failed to create macro storage directory: %v", err)
		return
	}
	data, err := json.MarshalIndent(s.macros, "", "  ")
	if err != nil {
		utils.Log.Error("Failed to marshal macros: %v", err)
		return
	}
	if err := os.WriteFile(s.storagePath, data, 0644); err != nil {
		utils.Log.Error("Failed to save macros: %v", err)
	}
}

// load 从文件加载宏
func (s *MacroService) load() {
	if s.storagePath == "" {
		return
	}
	data, err := os.ReadFile(s.storagePath)
	if err != nil {
		return
	}
	if err := json.Unmarshal(data, &s.macros); err != nil {
		utils.Log.Error("Failed to load macros: %v", err)
		s.macros = make([]Macro, 0)
	}
}

// 简易 ID 生成器
var (
	idCounter   uint64
	idCounterMu sync.Mutex
)

func generateID() string {
	now := time.Now().UnixNano()
	idCounterMu.Lock()
	idCounter++
	c := idCounter
	idCounterMu.Unlock()
	return fmt.Sprintf("%x-%x", now, c)
}
