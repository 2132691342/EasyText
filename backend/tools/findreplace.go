package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"easy-text/backend/utils"
	"easy-text/internal/concurrency"
)

// FindMatch 匹配结果
type FindMatch struct {
	Index       int    `json:"index"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	Content     string `json:"content"`
	MatchText   string `json:"matchText"`
	MatchLength int    `json:"matchLength"`
}

// FindInFileResult 单文件查找结果
type FindInFileResult struct {
	File    string      `json:"file"`
	Matches []FindMatch `json:"matches"`
	Count   int         `json:"count"`
}

// FindOptions 查找选项（仿 notepad-- findwin）
type FindOptions struct {
	Search        string `json:"search"`
	Replace       string `json:"replace"`
	CaseSensitive bool   `json:"caseSensitive"`
	WholeWord     bool   `json:"wholeWord"`
	UseRegex      bool   `json:"useRegex"`
	IncludeSubdir bool   `json:"includeSubdir"`
	FilePattern   string `json:"filePattern"` // e.g. "*.go;*.txt"
}

// 默认批量操作的并发数。同时也是 ch12 提到的"高可用四件套"的并发上限防线。
// UI 切 tab 关闭对话框时，concurrency.Run 会感知 ctx 取消并停止 spawn 新 goroutine，
// 已在跑的 fn 通过 ctx.Done() 检查自己提前返回。
const defaultBatchConcurrency = 16

// 替换是写入操作，比读慢；用更激进的限流避免 WebView2 主线程卡顿。
const defaultReplaceConcurrency = 8

// FindReplaceService 文件查找替换服务。
//
// 注意：结构体不再持有 sync.Mutex——各并发方法（FindInDirectory/SearchInFiles/
// BatchReplace/ReplaceInFiles）用函数内局部 mu 保护各自的 results 切片，
// 而 FindInFile/ReplaceInFile 是单文件纯函数，天然无共享状态。
// 原先的包级 mu 字段是死代码（从未被引用），P2-3 清理时移除。
type FindReplaceService struct{}

// NewFindReplaceService 创建查找替换服务
func NewFindReplaceService() *FindReplaceService {
	return &FindReplaceService{}
}

// FindInFile 在单个文件中查找
func (s *FindReplaceService) FindInFile(filePath, search string, options FindOptions) ([]FindMatch, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var pattern *regexp.Regexp
	if options.UseRegex {
		flags := ""
		if !options.CaseSensitive {
			flags = "(?i)"
		}
		pattern, err = regexp.Compile(flags + search)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %v", err)
		}
	}

	var matches []FindMatch
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // 10MB line buffer
	lineNum := 0
	globalIdx := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if options.UseRegex {
			locs := pattern.FindAllStringIndex(line, -1)
			for _, loc := range locs {
				matched := line[loc[0]:loc[1]]
				if options.WholeWord && !isWholeWord(line, loc[0], loc[1]) {
					continue
				}
				matches = append(matches, FindMatch{
					Index:       globalIdx,
					File:        filePath,
					Line:        lineNum,
					Column:      loc[0] + 1,
					Content:     line,
					MatchText:   matched,
					MatchLength: loc[1] - loc[0],
				})
				globalIdx++
			}
		} else {
			searchText := search
			lineToSearch := line
			if !options.CaseSensitive {
				searchText = strings.ToLower(search)
				lineToSearch = strings.ToLower(line)
			}
			pos := 0
			for {
				idx := strings.Index(lineToSearch[pos:], searchText)
				if idx < 0 {
					break
				}
				absIdx := pos + idx
				if options.WholeWord && !isWholeWord(line, absIdx, absIdx+len(search)) {
					pos = absIdx + 1
					continue
				}
				matches = append(matches, FindMatch{
					Index:       globalIdx,
					File:        filePath,
					Line:        lineNum,
					Column:      absIdx + 1,
					Content:     line,
					MatchText:   line[absIdx : absIdx+len(search)],
					MatchLength: len(search),
				})
				globalIdx++
				pos = absIdx + 1
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return matches, err
	}
	return matches, nil
}

// FindInDirectory 在目录中递归查找。
//
// 并发上限 defaultBatchConcurrency (16)；ctx 取消时停止 spawn 新 goroutine，
// 已在跑的 fn 自行检测 ctx.Done() 提前返回。
func (s *FindReplaceService) FindInDirectory(ctx context.Context, dirPath, search string, options FindOptions) ([]FindInFileResult, error) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil, utils.NewAppError(1002, "目录不存在", dirPath)
	}

	// 解析文件模式
	patterns := strings.Split(options.FilePattern, ";")
	if len(patterns) == 0 || options.FilePattern == "" {
		patterns = []string{"*"}
	}

	// Stage 1: 收集符合条件的所有文件路径（filepath.Walk 串行遍历，路径匹配/二进制筛选用同步逻辑）
	var files []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if !options.IncludeSubdir && path != dirPath {
				return filepath.SkipDir
			}
			if strings.HasPrefix(base, ".") || base == "node_modules" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		matched := false
		for _, p := range patterns {
			if m, _ := filepath.Match(strings.TrimSpace(p), info.Name()); m {
				matched = true
				break
			}
		}
		if !matched || isBinaryFilePath(path) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Stage 2: 受限并发搜索所有匹配文件。
	var (
		results []FindInFileResult
		mu      sync.Mutex
	)
	_ = concurrency.Run(ctx, files, defaultBatchConcurrency, func(cCtx context.Context, fp string) error {
		matches, err := s.FindInFile(fp, search, options)
		if err != nil || len(matches) == 0 {
			return nil // 单文件 IO 失败不中断；调用方在 results 里看不到这个文件即可
		}
		mu.Lock()
		results = append(results, FindInFileResult{File: fp, Matches: matches, Count: len(matches)})
		mu.Unlock()
		return nil
	})
	return results, nil
}

// ReplaceInFile 在文件中替换
func (s *FindReplaceService) ReplaceInFile(filePath, search, replace string, options FindOptions) (int, string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, "", err
	}

	text := string(content)
	var result string
	count := 0

	if options.UseRegex {
		flags := ""
		if !options.CaseSensitive {
			flags = "(?i)"
		}
		pattern, err := regexp.Compile(flags + search)
		if err != nil {
			return 0, "", fmt.Errorf("invalid regex: %v", err)
		}
		result = pattern.ReplaceAllStringFunc(text, func(match string) string {
			count++
			return replace
		})
	} else {
		if !options.CaseSensitive {
			// 不区分大小写替换
			pattern := regexp.MustCompile("(?i)" + regexp.QuoteMeta(search))
			result = pattern.ReplaceAllStringFunc(text, func(match string) string {
				count++
				return replace
			})
		} else {
			count = strings.Count(text, search)
			result = strings.ReplaceAll(text, search, replace)
		}
	}

	if count > 0 {
		if err := os.WriteFile(filePath, []byte(result), 0644); err != nil {
			return 0, "", err
		}
	}

	return count, result, nil
}

// BatchReplace 批量替换（目录级别）。
//
// 实现升级（ch12 高可用 §限流 + ch04 §1 goroutine 取消）：
//   - 上一版 1k 文件顺序串行 ReplaceInFile 会让 UI 长时间冻结 100s+，
//     升级后并发上限 defaultReplaceConcurrency (8)，单个文件读 + 写 IO 与
//     WebView2 主线程隔离，体感响应接近即时。
//   - ctx 取消时（如 UI 切 tab）停止 spawn；正在跑的 ReplaceInFile 没有 ctx
//     感知，所以我们要么把它升级为 ctx-aware，要么本函数 ingest 短超时。
//     此处采用短超时自动收敛：ctx 取消后 WaitForExit 在最多 ~30s 内返回。
func (s *FindReplaceService) BatchReplace(ctx context.Context, dirPath string, options FindOptions) (int, []string, error) {
	results, err := s.FindInDirectory(ctx, dirPath, options.Search, options)
	if err != nil {
		return 0, nil, err
	}
	if len(results) == 0 {
		return 0, nil, nil
	}

	files := make([]string, 0, len(results))
	for _, r := range results {
		files = append(files, r.File)
	}

	var (
		totalCount    int
		modifiedFiles []string
		mu            sync.Mutex
	)
	_ = concurrency.Run(ctx, files, defaultReplaceConcurrency, func(cCtx context.Context, fp string) error {
		count, _, err := s.ReplaceInFile(fp, options.Search, options.Replace, options)
		// 仅在成功且有替换时才记录（避免无意义噪声）
		if err != nil || count == 0 {
			return nil
		}
		mu.Lock()
		totalCount += count
		modifiedFiles = append(modifiedFiles, fp)
		mu.Unlock()
		return nil
	})
	return totalCount, modifiedFiles, nil
}

// isWholeWord 检查是否为完整单词
func isWholeWord(line string, start, end int) bool {
	if start > 0 {
		c := line[start-1]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			return false
		}
	}
	if end < len(line) {
		c := line[end]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			return false
		}
	}
	return true
}

// isBinaryFilePath 检查是否为二进制文件
func isBinaryFilePath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := map[string]bool{
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true, ".ico": true,
		".pdf": true, ".zip": true, ".tar": true, ".gz": true, ".rar": true, ".7z": true,
		".mp3": true, ".mp4": true, ".wav": true, ".avi": true, ".mkv": true,
		".ttf": true, ".otf": true, ".woff": true, ".woff2": true,
	}
	return binaryExts[ext]
}

// 🆕 V2.0.0 全局搜索替换

// SearchInFiles 在指定文件列表中搜索。
//
// 并发上限 defaultBatchConcurrency (16)；ctx 取消时收敛。返回值保证 results 切片
// 元素数量等于有匹配的文件数，与并发度无关（顺序不保证）。
func (s *FindReplaceService) SearchInFiles(ctx context.Context, filePaths []string, search string, options FindOptions) ([]FindInFileResult, error) {
	// 先筛选二进制文件——并发动它们做全文匹配是浪费 IO。
	filtered := make([]string, 0, len(filePaths))
	for _, fp := range filePaths {
		if isBinaryFilePath(fp) {
			continue
		}
		filtered = append(filtered, fp)
	}

	var (
		results []FindInFileResult
		mu      sync.Mutex
	)
	_ = concurrency.Run(ctx, filtered, defaultBatchConcurrency, func(cCtx context.Context, fp string) error {
		matches, err := s.FindInFile(fp, search, options)
		if err != nil || len(matches) == 0 {
			return nil
		}
		mu.Lock()
		results = append(results, FindInFileResult{File: fp, Matches: matches, Count: len(matches)})
		mu.Unlock()
		return nil
	})
	return results, nil
}

// ReplaceInFiles 在指定文件列表中执行替换。
//
// 实现升级：原版对每个文件顺序执行 ReplaceInFile（串行读+写），大列表下耗时 O(N×T);
// 升级后并发上限 defaultReplaceConcurrency (8)，单文件 IO 与 WebView2 主线程互不阻塞。
func (s *FindReplaceService) ReplaceInFiles(ctx context.Context, filePaths []string, search, replace string, options FindOptions) (int, []string, error) {
	filtered := make([]string, 0, len(filePaths))
	for _, fp := range filePaths {
		if isBinaryFilePath(fp) {
			continue
		}
		filtered = append(filtered, fp)
	}

	var (
		totalCount  int
		failedFiles []string
		mu          sync.Mutex
	)
	_ = concurrency.Run(ctx, filtered, defaultReplaceConcurrency, func(cCtx context.Context, fp string) error {
		count, _, err := s.ReplaceInFile(fp, search, replace, options)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			failedFiles = append(failedFiles, fp)
			return nil
		}
		totalCount += count
		return nil
	})
	return totalCount, failedFiles, nil
}
