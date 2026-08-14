package config

import (
	"github.com/glebarez/sqlite" // Pure Go SQLite driver for GORM
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database instance
var DB *gorm.DB

// Setting represents a user setting
type Setting struct {
	Key   string `gorm:"primaryKey;size:64;not null"`
	Value string `gorm:"size:4096"`
}

// InitDatabase initializes the SQLite database
func InitDatabase(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	// Auto migrate schemas
	err = DB.AutoMigrate(
		&Setting{},
		&Draft{},            // 🆕 V2.0.0
		&Snippet{},          // 🆕 V2.0.0
		&Bookmark{},         // 🆕 V2.0.0
		&RecentEntry{},      // 🆕 V2.0.0
		&RemoteConnection{}, // 🆕 V2.0.0
	)
	if err != nil {
		return err
	}

	return nil
}

// 🆕 V2.0.0 数据模型

// Draft 草稿表
type Draft struct {
	ID          int    `gorm:"primaryKey;autoIncrement"`
	FilePath    string `gorm:"uniqueIndex;not null;size:1024"`
	Content     string `gorm:"type:text"`
	Encoding    string `gorm:"default:utf-8;size:32"`
	LineEnding  string `gorm:"default:CRLF;size:8"`
	SavedAt     int64  `gorm:"autoCreateTime:milli"`
	FileModtime int64  `gorm:"default:0"`
}

// Snippet 代码片段表
type Snippet struct {
	ID          int    `gorm:"primaryKey;autoIncrement"`
	Name        string `gorm:"not null;size:256"`
	Prefix      string `gorm:"not null;size:128;uniqueIndex:idx_snippet_prefix_lang"`
	Body        string `gorm:"type:text;not null"`
	Description string `gorm:"default:'';size:512"`
	Language    string `gorm:"default:'';size:64;uniqueIndex:idx_snippet_prefix_lang"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt   int64  `gorm:"autoUpdateTime:milli"`
}

// Bookmark 书签表
type Bookmark struct {
	ID         int    `gorm:"primaryKey;autoIncrement"`
	FilePath   string `gorm:"not null;size:1024;index:idx_bookmark_file"`
	LineNumber int    `gorm:"not null"`
	Note       string `gorm:"default:'';size:512"`
	Tag        string `gorm:"default:'';size:64"`
	CreatedAt  int64  `gorm:"autoCreateTime:milli"`
}

// RecentEntry 最近访问记录表
type RecentEntry struct {
	ID         int    `gorm:"primaryKey;autoIncrement"`
	Path       string `gorm:"uniqueIndex;not null;size:1024"`
	IsFolder   bool   `gorm:"default:false;index:idx_recent_type"`
	Name       string `gorm:"not null;size:256"`
	AccessedAt int64  `gorm:"autoCreateTime:milli;index:idx_recent_time"`
}

// RemoteConnection 远程连接配置表（🆕 V2.0.0）
type RemoteConnection struct {
	ID                  string `gorm:"primaryKey;size:64"`
	Name                string `gorm:"not null;size:256"`
	Protocol            string `gorm:"not null;size:16"`
	Host                string `gorm:"not null;size:256"`
	Port                int    `gorm:"not null"`
	Username            string `gorm:"not null;size:128"`
	PasswordEncrypted   string `gorm:"size:512"`
	KeyFile             string `gorm:"size:1024"`
	PrivateKeyEncrypted string `gorm:"type:text"`
	RemotePath          string `gorm:"default:'/';size:1024"`
	CreatedAt           int64  `gorm:"autoCreateTime:milli"`
}

// GetSetting returns a setting value
func GetSetting(key string) (string, error) {
	var setting Setting
	result := DB.Where("key = ?", key).First(&setting)
	if result.Error == gorm.ErrRecordNotFound {
		return "", nil
	}
	return setting.Value, result.Error
}

// SetSetting sets a setting value
func SetSetting(key, value string) error {
	var setting Setting
	result := DB.Where("key = ?", key).First(&setting)

	if result.Error == gorm.ErrRecordNotFound {
		setting = Setting{Key: key, Value: value}
		return DB.Create(&setting).Error
	} else if result.Error != nil {
		return result.Error
	}

	return DB.Model(&setting).Update("value", value).Error
}
