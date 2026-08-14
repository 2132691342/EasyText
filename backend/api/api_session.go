package api

import (
	"encoding/json"
	"os"
	"path/filepath"

	"easy-text/backend/config"
)

// SessionFile 会话文件信息
type SessionFile struct {
	Path     string `json:"path"`
	Encoding string `json:"encoding"`
	Language string `json:"language"`
}

// Session 会话信息
type Session struct {
	Files    []SessionFile `json:"files"`
	ActiveID string        `json:"activeId"`
}

// SaveSession 保存当前会话（打开的文件列表）
func (h *Handler) SaveSession(files []SessionFile, activeID string) error {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}

	session := Session{
		Files:    files,
		ActiveID: activeID,
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(configDir, "session.json"), data, 0644)
}

// GetSession 获取上次会话信息
func (h *Handler) GetSession() (*Session, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	sessionPath := filepath.Join(configDir, "session.json")
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Session{}, nil
		}
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return &Session{}, nil
	}

	return &session, nil
}
