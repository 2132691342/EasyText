package api

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// RenameItem represents a file to be renamed
type RenameItem struct {
	OldPath string `json:"oldPath"`
	NewName string `json:"newName"`
	NewPath string `json:"newPath"`
	Status  string `json:"status"` // pending, success, error
	Error   string `json:"error,omitempty"`
}

// BatchRenameResult represents the result of a batch rename operation
type BatchRenameResult struct {
	Items   []RenameItem `json:"items"`
	Success int          `json:"success"`
	Failed  int          `json:"failed"`
	Total   int          `json:"total"`
}

// BatchRenamePreview generates a preview of renaming without actually renaming
// pattern can include: $N (original name), $E (extension), $I (index), $P (path)
// Examples: "file_$I.txt", "prefix_$N$E", "$N_backup$E"
func (h *Handler) BatchRenamePreview(files []string, pattern string, startIndex int, step int, delimiter string) *BatchRenameResult {
	result := &BatchRenameResult{
		Items: make([]RenameItem, 0, len(files)),
		Total: len(files),
	}

	// Sort files for consistent ordering
	sortedFiles := make([]string, len(files))
	copy(sortedFiles, files)
	sort.Strings(sortedFiles)

	for i, filePath := range sortedFiles {
		dir := filepath.Dir(filePath)
		ext := filepath.Ext(filePath)
		name := strings.TrimSuffix(filepath.Base(filePath), ext)

		index := startIndex + i*step
		newName := pattern
		newName = strings.ReplaceAll(newName, "$N", name)
		newName = strings.ReplaceAll(newName, "$E", ext)
		newName = strings.ReplaceAll(newName, "$I", strconv.Itoa(index))
		newName = strings.ReplaceAll(newName, "$P", dir)

		newPath := filepath.Join(dir, newName)

		result.Items = append(result.Items, RenameItem{
			OldPath: filePath,
			NewName: newName,
			NewPath: newPath,
			Status:  "pending",
		})
	}

	return result
}

// BatchRenameExecute executes a previewed rename operation
func (h *Handler) BatchRenameExecute(preview *BatchRenameResult) *BatchRenameResult {
	if preview == nil || len(preview.Items) == 0 {
		return preview
	}

	for i := range preview.Items {
		item := &preview.Items[i]
		if item.OldPath == item.NewPath {
			item.Status = "success"
			preview.Success++
			continue
		}

		// Check if target already exists
		if _, err := os.Stat(item.NewPath); err == nil {
			item.Status = "error"
			item.Error = fmt.Sprintf("目标文件已存在: %s", item.NewPath)
			preview.Failed++
			continue
		}

		// Ensure target directory exists
		targetDir := filepath.Dir(item.NewPath)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			item.Status = "error"
			item.Error = fmt.Sprintf("无法创建目标目录: %v", err)
			preview.Failed++
			continue
		}

		// Execute rename
		if err := os.Rename(item.OldPath, item.NewPath); err != nil {
			item.Status = "error"
			item.Error = fmt.Sprintf("重命名失败: %v", err)
			preview.Failed++
		} else {
			item.Status = "success"
			preview.Success++
		}
	}

	return preview
}

// GetCommonNamePatterns returns common renaming patterns
func (h *Handler) GetCommonNamePatterns() []string {
	return []string{
		"$N$E-$I",     // name.ext → name.ext-1
		"prefix_$I$E", // prefix_1.ext
		"$N$E",        // keep original
		"$N_backup$E", // name.ext → name_backup.ext
		"file_$I$E",   // file_1.ext
		"$N_$I$E",     // name_1.ext
	}
}
