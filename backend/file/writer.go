package file

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	"easy-text/backend/utils"

	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"
)

// FileWriter handles file writing operations
type FileWriter struct {
	chunkSize int // Chunk size for large file writing
}

// NewFileWriter creates a new FileWriter
func NewFileWriter() *FileWriter {
	return &FileWriter{chunkSize: 64 * 1024} // 64KB chunks
}

// WriteResult represents the result of writing a file
type WriteResult struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Success bool   `json:"success"`
}

// WriteFile writes content to a file
func (fw *FileWriter) WriteFile(path, content, encodingName string) (*WriteResult, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, utils.WrapError(1002, "无法创建目录", err)
	}

	// Get encoding
	var data []byte
	if encodingName != "" && encodingName != "UTF-8" {
		enc, err := ianaindex.IANA.Encoding(encodingName)
		if err != nil || enc == nil {
			data = []byte(content)
		} else {
			// Encode content
			encoder := enc.NewEncoder()
			encoded, err := io.ReadAll(transform.NewReader(bytes.NewReader([]byte(content)), encoder))
			if err != nil {
				return nil, utils.WrapError(3001, "编码转换失败", err)
			}
			data = encoded
		}
	} else {
		data = []byte(content)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return nil, utils.WrapError(1002, "无法写入文件", err)
	}

	// Get file size
	stat, _ := os.Stat(path)

	return &WriteResult{
		Path:    path,
		Size:    stat.Size(),
		Success: true,
	}, nil
}

// WriteFileWithBackup writes content to a file with backup
func (fw *FileWriter) WriteFileWithBackup(path, content, encodingName string) (*WriteResult, error) {
	// Create backup if file exists
	if _, err := os.Stat(path); err == nil {
		backupPath := path + ".bak"
		if err := os.Rename(path, backupPath); err != nil {
			// If rename fails, copy the file
			data, _ := os.ReadFile(path)
			if data != nil {
				os.WriteFile(backupPath, data, 0644)
			}
		}
	}

	return fw.WriteFile(path, content, encodingName)
}

// WriteLargeFile writes large content to a file in chunks
func (fw *FileWriter) WriteLargeFile(path, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return utils.WrapError(1002, "无法创建目录", err)
	}

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return utils.WrapError(1002, "无法创建文件", err)
	}
	defer file.Close()

	// Write in chunks
	data := []byte(content)
	for offset := 0; offset < len(data); offset += fw.chunkSize {
		end := offset + fw.chunkSize
		if end > len(data) {
			end = len(data)
		}
		if _, err := file.Write(data[offset:end]); err != nil {
			return utils.WrapError(1002, "无法写入文件", err)
		}
	}

	return nil
}

// AppendFile appends content to a file
func (fw *FileWriter) AppendFile(path, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return utils.WrapError(1002, "无法创建目录", err)
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return utils.WrapError(1002, "无法打开文件", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return utils.WrapError(1002, "无法写入文件", err)
	}

	return nil
}

// ConvertLineEnding converts line endings in content
func (fw *FileWriter) ConvertLineEnding(content, targetLineEnding string) string {
	// First normalize to LF
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	// Then convert to target
	switch targetLineEnding {
	case "CRLF":
		return strings.ReplaceAll(content, "\n", "\r\n")
	case "CR":
		return strings.ReplaceAll(content, "\n", "\r")
	default: // LF
		return content
	}
}

// DeleteFile deletes a file or empty directory
func DeleteFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return utils.ErrFileNotFound
	}
	return os.Remove(path)
}

// DeleteDirectory deletes a directory and all its contents
func DeleteDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return utils.ErrFileNotFound
	}
	return os.RemoveAll(path)
}

// RenameFile renames a file
func RenameFile(oldPath, newPath string) error {
	// Ensure target directory exists
	dir := filepath.Dir(newPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return utils.WrapError(1002, "无法创建目录", err)
	}
	return os.Rename(oldPath, newPath)
}

// CopyFile copies a file
func CopyFile(src, dst string) error {
	// Ensure target directory exists
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return utils.WrapError(1002, "无法创建目录", err)
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return utils.ErrFileNotFound
		}
		return utils.WrapError(1002, "无法打开源文件", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return utils.WrapError(1002, "无法创建目标文件", err)
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return utils.WrapError(1002, "无法复制文件", err)
	}

	// Copy permissions
	srcInfo, _ := srcFile.Stat()
	if srcInfo != nil {
		dstFile.Chmod(srcInfo.Mode())
	}

	return nil
}
