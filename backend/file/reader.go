package file

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"easy-text/backend/utils"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"
)

// FileInfo represents file information
type FileInfo struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	Ext        string `json:"ext"`
	Size       int64  `json:"size"`
	Modified   string `json:"modified"`
	Encoding   string `json:"encoding"`
	LineEnding string `json:"lineEnding"`
	LineCount  int    `json:"lineCount"`
	IsReadOnly bool   `json:"isReadOnly"`
	IsDir      bool   `json:"isDir"`
}

// ReadResult represents the result of reading a file
type ReadResult struct {
	Content          string   `json:"content"`
	Info             FileInfo `json:"info"`
	DetectedEncoding string   `json:"detectedEncoding"`
}

// FileReader handles file reading operations
type FileReader struct {
	maxFileSize int64 // Maximum file size in bytes
}

// 历史回顾：早期实现曾包含 ReadPartial / ReadJSONFile / ReadCSVFile / FileToString /
// FormatFileSize 等公开函数。S-7 清理时确认全项目无引用，删除以避免误用。
// 编码检测（DetectEncoding）已升级为 file.DetectEncoding（公开），与 tools.EncodingTool
// 的薄包装并存——后者标记为 Deprecated。
//
// 2026-08-14：ReadPartial 重新引入 —— 前端 LogViewer 大文件分块加载需要它。
// 实现为按行分块（见下方 ReadPartial），配套 file/reader_test.go 覆盖。

// NewFileReader creates a new FileReader
func NewFileReader(maxFileSize int64) *FileReader {
	if maxFileSize == 0 {
		maxFileSize = 100 * 1024 * 1024 // Default 100MB
	}
	return &FileReader{maxFileSize: maxFileSize}
}

// ReadFile reads a file and returns its content.
// Optimized hot path: UTF-8 files (95%+ of real-world usage) skip chardet entirely.
// Uses os.Open + stat + read to avoid the double-io of os.Stat + os.ReadFile.
func (fr *FileReader) ReadFile(path string) (*ReadResult, error) {
	// Open file once, then stat + read from the same handle
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, utils.ErrFileNotFound
		}
		return nil, utils.WrapError(1002, "无法访问文件", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, utils.WrapError(1002, "无法获取文件信息", err)
	}

	if stat.Size() > fr.maxFileSize {
		return nil, utils.ErrFileTooLarge
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, utils.WrapError(1002, "无法读取文件", err)
	}

	// Fast path: detect encoding. For UTF-8 (the common case), skip chardet.
	detectedEncoding, enc := fr.detectEncodingFast(data)

	// Decode content
	var content string
	if enc != nil {
		decoder := enc.NewDecoder()
		decoded, decErr := io.ReadAll(transform.NewReader(bytes.NewReader(data), decoder))
		if decErr != nil {
			content = string(data) // Fallback to raw bytes
		} else {
			content = string(decoded)
		}
	} else {
		content = string(data)
	}

	// One-pass metadata extraction: line-ending + line count
	lineEnding, lineCount := fr.detectLineEndingAndCount(content)

	// Read-only check: try opening with write flag
	isReadOnly := false
	if wf, err := os.OpenFile(path, os.O_WRONLY, 0); err != nil {
		isReadOnly = true
	} else {
		wf.Close()
	}

	info := FileInfo{
		Path:       path,
		Name:       filepath.Base(path),
		Ext:        strings.ToLower(strings.TrimPrefix(filepath.Ext(path), ".")),
		Size:       stat.Size(),
		Modified:   stat.ModTime().Format(time.RFC3339),
		Encoding:   detectedEncoding,
		LineEnding: lineEnding,
		LineCount:  lineCount,
		IsReadOnly: isReadOnly,
	}

	return &ReadResult{
		Content:          content,
		Info:             info,
		DetectedEncoding: detectedEncoding,
	}, nil
}

// DetectEncoding 检测字节流的编码。这是 file 包的权威实现，
// tools.EncodingTool.DetectEncoding 通过此函数复用，避免重复实现。
//
// 返回 (encodingName, encoding.Encoding)：
//   - name 是规范化后的 IANA 编码名（如 "UTF-8-BOM"、"GB18030"）
//   - enc 为 nil 表示 UTF-8（无需解码）
//
// Strategy:
// 1. Check BOM (instant).
// 2. Validate UTF-8 with utf8.Valid (microseconds, no allocation).
// 3. Only fall back to chardet for non-UTF-8 files.
func DetectEncoding(data []byte) (string, encoding.Encoding) {
	if len(data) == 0 {
		return "UTF-8", nil
	}

	// BOM check — UTF-8 BOM
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		enc, _ := ianaindex.IANA.Encoding("UTF-8")
		return "UTF-8-BOM", enc
	}

	// BOM check — UTF-16LE
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		enc, _ := ianaindex.IANA.Encoding("UTF-16LE")
		return "UTF-16LE", enc
	}

	// BOM check — UTF-16BE
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		enc, _ := ianaindex.IANA.Encoding("UTF-16BE")
		return "UTF-16BE", enc
	}

	// Fast path: utf8.Valid is O(n) with SIMD acceleration in Go 1.21+.
	// For the 95%+ of files that are UTF-8, this avoids chardet entirely.
	if utf8.Valid(data) {
		return "UTF-8", nil
	}

	// Slow path: non-UTF-8 file. Use chardet (statistical analysis).
	// Import moved to a separate file to keep this file clean.
	return detectEncodingWithChardet(data)
}

// detectEncodingWithChardet is the slow fallback for non-UTF-8 files.
// Kept separate so the fast path stays readable and the chardet import
// can be isolated if needed.
func detectEncodingWithChardet(data []byte) (string, encoding.Encoding) {
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(data)
	if err != nil {
		return "UTF-8", nil
	}

	encodingName := mapEncodingNameStatic(result.Charset)
	enc, err := ianaindex.IANA.Encoding(encodingName)
	if err != nil || enc == nil {
		return "UTF-8", nil
	}

	return encodingName, enc
}

// detectEncodingFast detects encoding with a fast path for UTF-8 (the common case).
// Kept as method on FileReader for existing callers; delegates to package-level
// DetectEncoding so logic lives in one place.
func (fr *FileReader) detectEncodingFast(data []byte) (string, encoding.Encoding) {
	return DetectEncoding(data)
}

// detectLineEndingAndCount detects line ending style and counts lines in one pass.
// Combined to avoid two separate traversals of the content string.
func (fr *FileReader) detectLineEndingAndCount(content string) (string, int) {
	if len(content) == 0 {
		return "LF", 1
	}

	var crlf, lfOnly, crOnly int
	lines := 1

	for i := 0; i < len(content); i++ {
		if content[i] == '\r' {
			if i+1 < len(content) && content[i+1] == '\n' {
				crlf++
				lines++
				i++ // skip \n
			} else {
				crOnly++
				lines++
			}
		} else if content[i] == '\n' {
			lfOnly++
			lines++
		}
	}

	switch {
	case crlf >= lfOnly && crlf >= crOnly:
		return "CRLF", lines
	case lfOnly >= crOnly:
		return "LF", lines
	default:
		return "CR", lines
	}
}

// mapEncodingNameStatic maps chardet encoding names to IANA names.
// Package-level function (no receiver needed).
func mapEncodingNameStatic(name string) string {
	mapping := map[string]string{
		"GB-18030":     "GB18030",
		"GB2312":       "GB18030",
		"GBK":          "GB18030",
		"ISO-8859-1":   "ISO-8859-1",
		"ISO-8859-2":   "ISO-8859-2",
		"UTF-8":        "UTF-8",
		"UTF-16LE":     "UTF-16LE",
		"UTF-16BE":     "UTF-16BE",
		"Big5":         "Big5",
		"Shift_JIS":    "Shift_JIS",
		"EUC-JP":       "EUC-JP",
		"EUC-KR":       "EUC-KR",
		"windows-1252": "Windows-1252",
		"windows-1251": "Windows-1251",
	}

	if mapped, ok := mapping[name]; ok {
		return mapped
	}
	return name
}

// ReadPartial 按行分块读取大文件：返回从第 offset 行（0-based）起最多 count 行的文本。
//
// 设计动机（ch12 高可用 §1 + P1-2 流式读取）：
//   - 一次性 io.ReadAll 加载 100MB 日志会让 WebView2 主线程冻结数秒；
//   - 分块按行读取让前端可以渐进渲染，用户先看到前 500 行，后台继续加载后续块。
//
// 语义：
//   - offset 为 0-based 行号；count 为要读取的最大行数。
//   - 返回文本不含行尾换行（逐行用 \n 连接），调用方自行 split。
//   - offset 超出总行数时返回空字符串，不报错。
//
// 注意：本实现用 bufio.Scanner 从文件头扫描到目标区间，复杂度 O(offset+count)。
// 对超深分页（offset 很大）不是最优，但日志文件行短，实际场景可接受；
// 若未来需要随机访问，可改成字节偏移 + 预建行索引。
func (fr *FileReader) ReadPartial(path string, offset, count int) (string, error) {
	if offset < 0 {
		offset = 0
	}
	if count <= 0 {
		return "", nil
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", utils.ErrFileNotFound
		}
		return "", utils.WrapError(1002, "无法访问文件", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // 单行最长 10MB

	var (
		lines   []string
		lineNum int
	)
	for scanner.Scan() {
		if lineNum >= offset && lineNum < offset+count {
			lines = append(lines, scanner.Text())
		}
		lineNum++
		if lineNum >= offset+count {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return "", utils.WrapError(1002, "无法读取文件", err)
	}
	return strings.Join(lines, "\n"), nil
}

// GetFileInfo returns file information without reading content
func (fr *FileReader) GetFileInfo(path string) (*FileInfo, error) {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, utils.ErrFileNotFound
	}
	if err != nil {
		return nil, utils.WrapError(1002, "无法访问文件", err)
	}

	// Check if read-only
	isReadOnly := false
	if file, err := os.OpenFile(path, os.O_WRONLY, 0644); err != nil {
		isReadOnly = true
	} else {
		file.Close()
	}

	return &FileInfo{
		Path:       path,
		Name:       filepath.Base(path),
		Ext:        strings.ToLower(strings.TrimPrefix(filepath.Ext(path), ".")),
		Size:       stat.Size(),
		Modified:   stat.ModTime().Format(time.RFC3339),
		IsReadOnly: isReadOnly,
	}, nil
}

// IsBinaryFile checks if a file is binary
func (fr *FileReader) IsBinaryFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, utils.WrapError(1002, "无法打开文件", err)
	}
	defer file.Close()

	// Read first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false, utils.WrapError(1002, "无法读取文件", err)
	}

	// Check for null bytes (common in binary files)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, nil
		}
	}

	return false, nil
}
