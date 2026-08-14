package file

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"easy-text/backend/utils"
)

// TreeNode represents a node in the file tree
type TreeNode struct {
	Path     string      `json:"path"`
	Name     string      `json:"name"`
	IsDir    bool        `json:"isDir"`
	Children []*TreeNode `json:"children,omitempty"`
	Expanded bool        `json:"expanded"`
	Ext      string      `json:"ext,omitempty"`
}

// FileTree represents a file tree structure
type FileTree struct {
	Root     *TreeNode `json:"root"`
	BasePath string    `json:"basePath"`
}

// TreeBuilder builds file tree structures
type TreeBuilder struct {
	ignorePatterns []string
	maxDepth       int
}

// NewTreeBuilder creates a new TreeBuilder.
// A text editor should show ALL files; only system-level VCS/metadata dirs are hidden.
func NewTreeBuilder() *TreeBuilder {
	return &TreeBuilder{
		ignorePatterns: []string{
			".git",  // VCS metadata – not user files
			"~$*",   // Office temp lock files (Excel/Word/PowerPoint)
			"*.tmp", // Generic temp files
		},
		maxDepth: 20,
	}
}

// BuildTree builds a file tree from a directory path.
// Returns a tree even if some subdirectories cannot be read (they are skipped).
func (tb *TreeBuilder) BuildTree(path string) (*FileTree, error) {
	// Check if path exists
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, utils.ErrDirectoryNotFound
	}
	if err != nil {
		return nil, utils.WrapError(1002, "无法访问目录", err)
	}

	// Create root node
	root := &TreeNode{
		Path:     path,
		Name:     filepath.Base(path),
		IsDir:    stat.IsDir(),
		Expanded: true,
	}

	if stat.IsDir() {
		// Errors reading the root directory are fatal;
		// errors inside subdirectories are skipped (continue).
		tb.buildDirectoryTree(root, 0)
	}

	return &FileTree{
		Root:     root,
		BasePath: path,
	}, nil
}

// buildDirectoryTree recursively builds tree for a directory.
// Unreadable subdirectories are silently skipped.
func (tb *TreeBuilder) buildDirectoryTree(node *TreeNode, depth int) {
	if depth >= tb.maxDepth {
		return
	}

	// Read directory entries
	entries, err := os.ReadDir(node.Path)
	if err != nil {
		// Cannot read this directory – skip it
		utils.Log.Debug("Cannot read directory %s: %v", node.Path, err)
		return
	}

	// Sort entries: directories first, then files
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	// Process entries
	for _, entry := range entries {
		if tb.shouldIgnore(entry.Name()) {
			continue
		}

		childPath := filepath.Join(node.Path, entry.Name())
		child := &TreeNode{
			Path:  childPath,
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
		}

		if !entry.IsDir() {
			child.Ext = strings.ToLower(strings.TrimPrefix(filepath.Ext(entry.Name()), "."))
		}

		node.Children = append(node.Children, child)

		// Recursively build subdirectories
		if entry.IsDir() {
			tb.buildDirectoryTree(child, depth+1)
		}
	}
}

// shouldIgnore checks if a name should be ignored
func (tb *TreeBuilder) shouldIgnore(name string) bool {
	for _, pattern := range tb.ignorePatterns {
		switch {
		case strings.HasSuffix(pattern, "*") && !strings.Contains(pattern[:len(pattern)-1], "*"):
			// Prefix wildcard: "~$*" matches names starting with "~$"
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(name, prefix) {
				return true
			}
		case strings.HasPrefix(pattern, "*") && !strings.Contains(pattern[1:], "*"):
			// Suffix wildcard: "*.tmp" matches names ending with ".tmp"
			suffix := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(name, suffix) {
				return true
			}
		default:
			// Exact match
			if name == pattern {
				return true
			}
		}
	}
	return false
}

// ExpandNode expands a tree node
func (ft *FileTree) ExpandNode(path string) {
	node := ft.findNode(path)
	if node != nil && node.IsDir {
		node.Expanded = true
	}
}

// CollapseNode collapses a tree node
func (ft *FileTree) CollapseNode(path string) {
	node := ft.findNode(path)
	if node != nil && node.IsDir {
		node.Expanded = false
	}
}

// findNode finds a node by path
func (ft *FileTree) findNode(path string) *TreeNode {
	return findNodeRecursive(ft.Root, path)
}

func findNodeRecursive(node *TreeNode, path string) *TreeNode {
	if node.Path == path {
		return node
	}
	for _, child := range node.Children {
		if strings.HasPrefix(path, child.Path) {
			result := findNodeRecursive(child, path)
			if result != nil {
				return result
			}
		}
	}
	return nil
}

// GetFileList returns a flat list of files
func (ft *FileTree) GetFileList() []TreeNode {
	return getFileListRecursive(ft.Root)
}

func getFileListRecursive(node *TreeNode) []TreeNode {
	var files []TreeNode
	if !node.IsDir {
		files = append(files, *node)
	}
	for _, child := range node.Children {
		files = append(files, getFileListRecursive(child)...)
	}
	return files
}

// GetDirectoryList returns a flat list of directories
func (ft *FileTree) GetDirectoryList() []TreeNode {
	return getDirListRecursive(ft.Root)
}

func getDirListRecursive(node *TreeNode) []TreeNode {
	var dirs []TreeNode
	if node.IsDir {
		dirs = append(dirs, *node)
		for _, child := range node.Children {
			if child.IsDir {
				dirs = append(dirs, getDirListRecursive(child)...)
			}
		}
	}
	return dirs
}

// SetIgnorePatterns sets custom ignore patterns
func (tb *TreeBuilder) SetIgnorePatterns(patterns []string) {
	tb.ignorePatterns = patterns
}

// SetMaxDepth sets the maximum tree depth
func (tb *TreeBuilder) SetMaxDepth(depth int) {
	tb.maxDepth = depth
}
