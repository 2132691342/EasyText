package tools

import (
	"context"
	"strings"
	"sync"

	"easy-text/backend/utils"

	dmp "github.com/sergi/go-diff/diffmatchpatch"
)

// DiffResult represents the result of a diff operation
type DiffResult struct {
	Diffs     []DiffItem `json:"diffs"`
	Added     int        `json:"added"`
	Removed   int        `json:"removed"`
	Modified  int        `json:"modified"`
	Unchanged int        `json:"unchanged"`
}

// DiffItem represents a single diff item
type DiffItem struct {
	Type     string `json:"type"` // added, removed, modified, unchanged
	Content  string `json:"content"`
	OldLine  int    `json:"oldLine,omitempty"`
	NewLine  int    `json:"newLine,omitempty"`
	Position int    `json:"position"`
}

// DiffLine represents a line in diff view
type DiffLine struct {
	Type      string `json:"type"` // added, removed, modified, unchanged
	Content   string `json:"content"`
	OldLine   int    `json:"oldLine"`
	NewLine   int    `json:"newLine"`
	DiffChars string `json:"diffChars,omitempty"` // character-level diff
}

// DiffBlock represents a block of diff lines
type DiffBlock struct {
	OldStart int        `json:"oldStart"`
	OldCount int        `json:"oldCount"`
	NewStart int        `json:"newStart"`
	NewCount int        `json:"newCount"`
	Lines    []DiffLine `json:"lines"`
}

// DiffTool provides diff comparison functions
type DiffTool struct {
	dmp *dmp.DiffMatchPatch
}

// NewDiffTool creates a new DiffTool
func NewDiffTool() *DiffTool {
	return &DiffTool{
		dmp: dmp.New(),
	}
}

// Compare compares two texts and returns diffs
func (dt *DiffTool) Compare(oldText, newText string) (*DiffResult, error) {
	diffs := dt.dmp.DiffMain(oldText, newText, true)

	result := &DiffResult{
		Diffs: make([]DiffItem, 0, len(diffs)),
	}

	position := 0
	for _, d := range diffs {
		item := DiffItem{
			Position: position,
			Content:  d.Text,
		}

		switch d.Type {
		case dmp.DiffInsert:
			item.Type = "added"
			result.Added += countLines(d.Text)
		case dmp.DiffDelete:
			item.Type = "removed"
			result.Removed += countLines(d.Text)
		case dmp.DiffEqual:
			item.Type = "unchanged"
			result.Unchanged += countLines(d.Text)
		}

		result.Diffs = append(result.Diffs, item)
		position += len(d.Text)
	}

	return result, nil
}

// CompareLines compares two texts line by line
func (dt *DiffTool) CompareLines(oldText, newText string) ([]DiffBlock, error) {
	oldLines := splitLines(oldText)
	newLines := splitLines(newText)

	// Calculate diff using LCS algorithm
	lcs := dt.calculateLCS(oldLines, newLines)

	// Build diff blocks
	blocks := dt.buildDiffBlocks(oldLines, newLines, lcs)

	return blocks, nil
}

// calculateLCS calculates the Longest Common Subsequence
func (dt *DiffTool) calculateLCS(oldLines, newLines []string) [][]bool {
	m := len(oldLines)
	n := len(newLines)

	// Create LCS table
	lcs := make([][]int, m+1)
	for i := range lcs {
		lcs[i] = make([]int, n+1)
	}

	// Fill LCS table
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if oldLines[i-1] == newLines[j-1] {
				lcs[i][j] = lcs[i-1][j-1] + 1
			} else if lcs[i-1][j] >= lcs[i][j-1] {
				lcs[i][j] = lcs[i-1][j]
			} else {
				lcs[i][j] = lcs[i][j-1]
			}
		}
	}

	// Build match matrix
	matches := make([][]bool, m)
	for i := range matches {
		matches[i] = make([]bool, n)
	}

	// Trace back to find matches
	i, j := m, n
	for i > 0 && j > 0 {
		if oldLines[i-1] == newLines[j-1] {
			matches[i-1][j-1] = true
			i--
			j--
		} else if lcs[i-1][j] >= lcs[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return matches
}

// buildDiffBlocks builds diff blocks from matched lines
func (dt *DiffTool) buildDiffBlocks(oldLines, newLines []string, matches [][]bool) []DiffBlock {
	var blocks []DiffBlock

	oldIdx := 0
	newIdx := 0

	for oldIdx < len(oldLines) || newIdx < len(newLines) {
		// Find next match or end
		nextOldMatch := -1
		nextNewMatch := -1

		for i := oldIdx; i < len(oldLines); i++ {
			for j := newIdx; j < len(newLines); j++ {
				if i < len(matches) && j < len(matches[i]) && matches[i][j] {
					nextOldMatch = i
					nextNewMatch = j
					break
				}
			}
			if nextOldMatch >= 0 {
				break
			}
		}

		// If no match found, create block for remaining lines
		if nextOldMatch < 0 {
			if oldIdx < len(oldLines) || newIdx < len(newLines) {
				block := DiffBlock{
					OldStart: oldIdx + 1,
					OldCount: len(oldLines) - oldIdx,
					NewStart: newIdx + 1,
					NewCount: len(newLines) - newIdx,
					Lines:    make([]DiffLine, 0),
				}

				// Add remaining old lines as removed
				for i := oldIdx; i < len(oldLines); i++ {
					block.Lines = append(block.Lines, DiffLine{
						Type:    "removed",
						Content: oldLines[i],
						OldLine: i + 1,
						NewLine: 0,
					})
				}

				// Add remaining new lines as added
				for j := newIdx; j < len(newLines); j++ {
					block.Lines = append(block.Lines, DiffLine{
						Type:    "added",
						Content: newLines[j],
						OldLine: 0,
						NewLine: j + 1,
					})
				}

				blocks = append(blocks, block)
			}
			break
		}

		// Create block for changes before match
		if nextOldMatch > oldIdx || nextNewMatch > newIdx {
			block := DiffBlock{
				OldStart: oldIdx + 1,
				OldCount: nextOldMatch - oldIdx,
				NewStart: newIdx + 1,
				NewCount: nextNewMatch - newIdx,
				Lines:    make([]DiffLine, 0),
			}

			// Add old lines as removed
			for i := oldIdx; i < nextOldMatch; i++ {
				block.Lines = append(block.Lines, DiffLine{
					Type:    "removed",
					Content: oldLines[i],
					OldLine: i + 1,
					NewLine: 0,
				})
			}

			// Add new lines as added
			for j := newIdx; j < nextNewMatch; j++ {
				block.Lines = append(block.Lines, DiffLine{
					Type:    "added",
					Content: newLines[j],
					OldLine: 0,
					NewLine: j + 1,
				})
			}

			blocks = append(blocks, block)
		}

		// Create block for matched line
		block := DiffBlock{
			OldStart: nextOldMatch + 1,
			OldCount: 1,
			NewStart: nextNewMatch + 1,
			NewCount: 1,
			Lines: []DiffLine{
				{
					Type:    "unchanged",
					Content: oldLines[nextOldMatch],
					OldLine: nextOldMatch + 1,
					NewLine: nextNewMatch + 1,
				},
			},
		}
		blocks = append(blocks, block)

		oldIdx = nextOldMatch + 1
		newIdx = nextNewMatch + 1
	}

	return blocks
}

// CompareCharacters compares two strings character by character
func (dt *DiffTool) CompareCharacters(oldStr, newStr string) string {
	diffs := dt.dmp.DiffMain(oldStr, newStr, true)

	var result strings.Builder
	for _, d := range diffs {
		switch d.Type {
		case dmp.DiffInsert:
			result.WriteString("<ins>" + d.Text + "</ins>")
		case dmp.DiffDelete:
			result.WriteString("<del>" + d.Text + "</del>")
		case dmp.DiffEqual:
			result.WriteString(d.Text)
		}
	}

	return result.String()
}

// CompareLargeFiles 并发对比两个大文本。chunkSize 给定时按等长切片并发计算。
//
// 实现升级（ch04 §1 + ch12 §高可用）：
//   - 上版裸 goroutine 数量 O(N)，N = max(oldChunks, newChunks)。
//     N 大时同时打开 N 个 goroutine 会把 WebView2 主线程拖到卡。
//   - 升级后用 internal/concurrency.Run 限制 16 个并发；ctx 取消时停止 spawn。
//   - chunkResults 按 idx 写，不存在数据竞争——每个 goroutine 只写自己槽位。
func (dt *DiffTool) CompareLargeFiles(ctx context.Context, oldText, newText string, chunkSize int) (*DiffResult, error) {
	oldChunks := splitChunks(oldText, chunkSize)
	newChunks := splitChunks(newText, chunkSize)
	n := max(len(oldChunks), len(newChunks))
	chunkResults := make([]DiffResult, n)

	// 把 idx + 各自的 chunk 打包成结构体让 parallelism.Run 调度；
	// 这里仍用 Wg + channel sem 是因为每个 idx 写自己的槽位，ctx 取消也能早退。
	sem := make(chan struct{}, defaultBatchConcurrency)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		if ctx.Err() != nil {
			break
		}
		idx := i
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			var oldChunk, newChunk string
			if idx < len(oldChunks) {
				oldChunk = oldChunks[idx]
			}
			if idx < len(newChunks) {
				newChunk = newChunks[idx]
			}
			result, err := dt.Compare(oldChunk, newChunk)
			if err == nil {
				chunkResults[idx] = *result
			}
		}()
	}
	wg.Wait()

	// Merge results
	finalResult := &DiffResult{Diffs: make([]DiffItem, 0)}
	for _, cr := range chunkResults {
		finalResult.Diffs = append(finalResult.Diffs, cr.Diffs...)
		finalResult.Added += cr.Added
		finalResult.Removed += cr.Removed
		finalResult.Unchanged += cr.Unchanged
	}
	return finalResult, nil
}

// GetPatch generates a patch for diff
func (dt *DiffTool) GetPatch(oldText, newText string) string {
	diffs := dt.dmp.DiffMain(oldText, newText, true)
	patches := dt.dmp.PatchMake(diffs)
	return dt.dmp.PatchToText(patches)
}

// ApplyPatch applies a patch to text
func (dt *DiffTool) ApplyPatch(text, patch string) (string, error) {
	patches, err := dt.dmp.PatchFromText(patch)
	if err != nil {
		return "", utils.WrapError(4001, "无法解析补丁", err)
	}

	result, _ := dt.dmp.PatchApply(patches, text)
	return result, nil
}

// Helper functions

func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}

func splitChunks(s string, size int) []string {
	if size <= 0 {
		return []string{s}
	}

	var chunks []string
	for i := 0; i < len(s); i += size {
		end := i + size
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

// SemanticDiff performs semantic diff (ignoring whitespace changes)
func (dt *DiffTool) SemanticDiff(oldText, newText string) (*DiffResult, error) {
	// Normalize whitespace
	normalizedOld := normalizeWhitespace(oldText)
	normalizedNew := normalizeWhitespace(newText)

	return dt.Compare(normalizedOld, normalizedNew)
}

func normalizeWhitespace(s string) string {
	// Replace multiple spaces with single space
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.Join(strings.Fields(s), " ")
	return s
}
