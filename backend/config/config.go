package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"easy-text/backend/utils"
)

// AppConfig represents the application configuration
type AppConfig struct {
	Version int          `json:"version"` // Schema version for compatibility
	Editor  EditorConfig `json:"editor"`
	Theme   ThemeConfig  `json:"theme"`
	File    FileConfig   `json:"file"`
	UI      UIConfig     `json:"ui"`
}

// EditorConfig represents editor settings
type EditorConfig struct {
	FontSize         int    `json:"fontSize"`
	FontFamily       string `json:"fontFamily"`
	TabSize          int    `json:"tabSize"`
	InsertSpaces     bool   `json:"insertSpaces"`
	WordWrap         bool   `json:"wordWrap"`
	LineNumbers      bool   `json:"lineNumbers"`
	AutoSave         bool   `json:"autoSave"`
	AutoSaveInterval int    `json:"autoSaveInterval"` // seconds
	HighlightLine    bool   `json:"highlightLine"`
	BracketPairColor bool   `json:"bracketPairColor"`
	Minimap          bool   `json:"minimap"`
	// 视图显示开关（对齐 notepad-- 视图菜单）
	ShowIndentGuide bool `json:"showIndentGuide"`
	ShowWhitespace  bool `json:"showWhitespace"`
	ShowEol         bool `json:"showEol"`
	FoldEnable      bool `json:"foldEnable"`
	// 列块模式开关
	ColumnMode bool `json:"columnMode"`
	// 🆕 V2.0.0 列块编辑配置
	ColumnModeConfig ColumnModeConfig `json:"columnModeConfig"`
}

// ColumnModeConfig 列块编辑配置（🆕 V2.0.0）
type ColumnModeConfig struct {
	NumberStart    int    `json:"numberStart"`    // 数字序列起始值
	NumberStep     int    `json:"numberStep"`     // 数字序列步长
	NumberBase     int    `json:"numberBase"`     // 进制 2/8/10/16
	DateFormat     string `json:"dateFormat"`     // 日期格式
	CaseConversion string `json:"caseConversion"` // 大小写转换类型
}

// ThemeConfig represents theme settings
type ThemeConfig struct {
	CurrentTheme string `json:"currentTheme"`
	AutoTheme    bool   `json:"autoTheme"` // 🆕 V2.0.0 自动跟随系统主题
}

// FileConfig represents file settings
type FileConfig struct {
	DefaultEncoding    string   `json:"defaultEncoding"`
	AutoDetectEncoding bool     `json:"autoDetectEncoding"`
	DefaultLineEnding  string   `json:"defaultLineEnding"`
	IgnorePatterns     []string `json:"ignorePatterns"` // 🆕 V2.0.0 文件过滤规则
}

// UIConfig represents UI settings
type UIConfig struct {
	Language         string          `json:"language"`
	ShowFileTree     bool            `json:"showFileTree"`
	ShowStatusBar    bool            `json:"showStatusBar"`
	ShowToolBar      bool            `json:"showToolBar"`
	ShowFileListView bool            `json:"showFileListView"`
	ShowWebAddr      bool            `json:"showWebAddr"`
	FileTreeWidth    int             `json:"fileTreeWidth"`
	ZoomLevel        int             `json:"zoomLevel"`
	ToolbarIconSize  int             `json:"toolbarIconSize"`  // 24 | 36 | 48
	Favorites        []string        `json:"favorites"`        // 收藏夹文件路径列表
	StatusBarItems   map[string]bool `json:"statusBarItems"`   // 🆕 V2.0.0 状态栏显示项
	ToolbarItems     map[string]bool `json:"toolbarItems"`     // 🆕 V2.0.0 工具栏显示项
	RecentFilesLimit int             `json:"recentFilesLimit"` // 🆕 V2.0.0 最近文件数限制
	LastFolder       string          `json:"lastFolder"`       // 🆕 V2.0.0 上次打开的项目目录
	// 🆕 关闭行为：true 时关闭按钮最小化到托盘；false 时直接退出
	CloseToTray bool `json:"closeToTray"`
}

// Default configuration values
var defaultConfig = AppConfig{
	Version: 3,
	Editor: EditorConfig{
		FontSize:         14,
		FontFamily:       "Consolas, 'Courier New', monospace",
		TabSize:          4,
		InsertSpaces:     true,
		WordWrap:         false,
		LineNumbers:      true,
		AutoSave:         false,
		AutoSaveInterval: 60,
		HighlightLine:    true,
		BracketPairColor: true,
		Minimap:          true,
		ShowIndentGuide:  true,
		ShowWhitespace:   false,
		ShowEol:          false,
		FoldEnable:       true,
		ColumnMode:       false,
		ColumnModeConfig: ColumnModeConfig{
			NumberStart: 1,
			NumberStep:  1,
			NumberBase:  10,
			DateFormat:  "2006-01-02",
		},
	},
	Theme: ThemeConfig{
		CurrentTheme: "light",
		AutoTheme:    false,
	},
	File: FileConfig{
		DefaultEncoding:    "UTF-8",
		AutoDetectEncoding: true,
		DefaultLineEnding:  "CRLF",
		IgnorePatterns:     []string{".git", "~$*", "*.tmp", "node_modules"},
	},
	UI: UIConfig{
		Language:         "zh-CN",
		ShowFileTree:     true,
		ShowStatusBar:    true,
		ShowToolBar:      true,
		ShowFileListView: false,
		ShowWebAddr:      false,
		FileTreeWidth:    250,
		ZoomLevel:        100,
		ToolbarIconSize:  24,
		Favorites:        []string{},
		StatusBarItems: map[string]bool{
			"cursor": true, "lines": true, "encoding": true,
			"lineEnding": true, "language": true, "path": true, "zoom": true,
		},
		ToolbarItems:     map[string]bool{},
		RecentFilesLimit: 10,
		// 默认开托盘驻留：防止误关闭丢失正在编辑的内容；用户可自行关闭。
		CloseToTray: true,
	},
}

// ConfigManager manages application configuration
type ConfigManager struct {
	config     AppConfig
	configPath string
	mu         sync.RWMutex
}

// Global config manager instance
var Config *ConfigManager

// InitConfig initializes the configuration manager
func InitConfig(configPath string) error {
	Config = &ConfigManager{
		configPath: configPath,
		config:     defaultConfig,
	}
	return Config.Load()
}

// Load loads configuration from file
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if config file exists
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// Create default config
		return cm.saveWithoutLock()
	}

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return err
	}

	// Parse JSON and merge with defaults
	var loadedConfig AppConfig
	if err := json.Unmarshal(data, &loadedConfig); err != nil {
		return err
	}

	// Merge with defaults (loaded values take precedence)
	cm.config = mergeConfig(defaultConfig, loadedConfig)

	return nil
}

// Save saves configuration to file
func (cm *ConfigManager) Save() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.saveWithoutLock()
}

func (cm *ConfigManager) saveWithoutLock() error {
	// Ensure directory exists
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.configPath, data, 0644)
}

// Get returns the current configuration
func (cm *ConfigManager) Get() AppConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// Update updates the configuration
func (cm *ConfigManager) Update(newConfig AppConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config = newConfig
	return cm.saveWithoutLock()
}

// UpdateEditor updates editor settings
func (cm *ConfigManager) UpdateEditor(editor EditorConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config.Editor = editor
	return cm.saveWithoutLock()
}

// UpdateTheme updates theme settings
func (cm *ConfigManager) UpdateTheme(theme ThemeConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config.Theme = theme
	return cm.saveWithoutLock()
}

// UpdateUI updates UI settings
func (cm *ConfigManager) UpdateUI(ui UIConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config.UI = ui
	return cm.saveWithoutLock()
}

// mergeConfig merges loaded config with defaults.
// version >= 3: current schema, trusted as-is (missing fields use Go zero values,
// which for the new bool/slice/map fields is a safe default).
// version 2: migrate to v3 — fill in new fields from defaults.
// version 1 (legacy): trusted for known fields, but new v2/v3 fields fall back to defaults.
// version 0/missing: full field-by-field merge to resolve bool ambiguity.
func mergeConfig(defaults, loaded AppConfig) AppConfig {
	// version >= 3: current schema
	if loaded.Version >= 3 {
		loaded.Version = defaults.Version // always use latest version
		// Ensure nil slices/maps
		if loaded.UI.Favorites == nil {
			loaded.UI.Favorites = []string{}
		}
		if loaded.UI.StatusBarItems == nil {
			loaded.UI.StatusBarItems = defaults.UI.StatusBarItems
		}
		if loaded.UI.ToolbarItems == nil {
			loaded.UI.ToolbarItems = defaults.UI.ToolbarItems
		}
		if loaded.File.IgnorePatterns == nil {
			loaded.File.IgnorePatterns = defaults.File.IgnorePatterns
		}
		if loaded.UI.RecentFilesLimit == 0 {
			loaded.UI.RecentFilesLimit = defaults.UI.RecentFilesLimit
		}
		// 兼容 V3：旧版本 JSON 没有 closeToTray 字段，使用默认值（开托盘驻留）
		_ = loaded.UI.CloseToTray
		return loaded
	}

	// v2 → v3 迁移
	if loaded.Version >= 2 {
		loaded.Version = defaults.Version
		if loaded.UI.Favorites == nil {
			loaded.UI.Favorites = []string{}
		}
		// 填充 v3 新增字段
		loaded.Theme.AutoTheme = defaults.Theme.AutoTheme
		loaded.File.IgnorePatterns = defaults.File.IgnorePatterns
		loaded.UI.StatusBarItems = defaults.UI.StatusBarItems
		loaded.UI.ToolbarItems = defaults.UI.ToolbarItems
		loaded.UI.RecentFilesLimit = defaults.UI.RecentFilesLimit
		loaded.Editor.ColumnModeConfig = defaults.Editor.ColumnModeConfig
		// V2 配置文件 JSON 缺少 CloseToTray 字段，反序列化得到零值 (false)。
		// 这里强制覆盖为默认值，保持"默认开托盘"的升级体验一致。
		loaded.UI.CloseToTray = defaults.UI.CloseToTray
		return loaded
	}

	// Legacy config (version 0 or 1): merge with defaults
	result := defaults

	// Editor
	if loaded.Editor.FontSize > 0 {
		result.Editor.FontSize = loaded.Editor.FontSize
	}
	if loaded.Editor.FontFamily != "" {
		result.Editor.FontFamily = loaded.Editor.FontFamily
	}
	if loaded.Editor.TabSize > 0 {
		result.Editor.TabSize = loaded.Editor.TabSize
	}
	result.Editor.InsertSpaces = loaded.Editor.InsertSpaces
	result.Editor.WordWrap = loaded.Editor.WordWrap
	result.Editor.LineNumbers = loaded.Editor.LineNumbers
	result.Editor.AutoSave = loaded.Editor.AutoSave
	if loaded.Editor.AutoSaveInterval > 0 {
		result.Editor.AutoSaveInterval = loaded.Editor.AutoSaveInterval
	}
	result.Editor.HighlightLine = loaded.Editor.HighlightLine
	result.Editor.BracketPairColor = loaded.Editor.BracketPairColor
	result.Editor.Minimap = loaded.Editor.Minimap

	// Theme
	if loaded.Theme.CurrentTheme != "" {
		result.Theme.CurrentTheme = loaded.Theme.CurrentTheme
	}

	// File
	if loaded.File.DefaultEncoding != "" {
		result.File.DefaultEncoding = loaded.File.DefaultEncoding
	}
	result.File.AutoDetectEncoding = loaded.File.AutoDetectEncoding
	if loaded.File.DefaultLineEnding != "" {
		result.File.DefaultLineEnding = loaded.File.DefaultLineEnding
	}

	// UI
	if loaded.UI.Language != "" {
		result.UI.Language = loaded.UI.Language
	}
	result.UI.ShowFileTree = loaded.UI.ShowFileTree
	result.UI.ShowStatusBar = loaded.UI.ShowStatusBar
	if loaded.UI.FileTreeWidth > 0 {
		result.UI.FileTreeWidth = loaded.UI.FileTreeWidth
	}
	if loaded.UI.ZoomLevel > 0 {
		result.UI.ZoomLevel = loaded.UI.ZoomLevel
	}

	return result
}

// GetConfigDir returns the configuration directory.
// 便携模式下，返回程序目录下的 data/ 文件夹；否则使用 os.UserConfigDir()。
func GetConfigDir() (string, error) {
	// 便携模式检测
	if IsPortableMode() {
		exePath, err := os.Executable()
		if err != nil {
			return "", utils.WrapError(5001, "无法获取程序路径", err)
		}
		dataDir := filepath.Join(filepath.Dir(exePath), "data")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return "", utils.WrapError(5001, "无法创建便携版数据目录", err)
		}
		return dataDir, nil
	}

	baseDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback for older systems
		baseDir = os.Getenv("LOCALAPPDATA")
		if baseDir == "" {
			baseDir = os.Getenv("HOME")
			if baseDir == "" {
				return "", utils.WrapError(5001, "无法确定配置目录", err)
			}
			baseDir = filepath.Join(baseDir, ".config")
		}
	}
	configDir := filepath.Join(baseDir, "EasyText")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", utils.WrapError(5001, "无法创建配置目录", err)
	}
	return configDir, nil
}

// IsPortableMode 检查程序同级目录是否存在 portable.dat 标记文件
func IsPortableMode() bool {
	exePath, err := os.Executable()
	if err != nil {
		return false
	}
	exeDir := filepath.Dir(exePath)
	portableFile := filepath.Join(exeDir, "portable.dat")
	_, err = os.Stat(portableFile)
	return err == nil
}

// GetDataDir returns the data directory
func GetDataDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	dataDir := filepath.Join(configDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", utils.WrapError(5001, "无法创建数据目录", err)
	}
	return dataDir, nil
}
