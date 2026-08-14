package file

import (
	"os"
	"path/filepath"
	"sync"

	"easy-text/backend/utils"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher watches for file system changes
type FileWatcher struct {
	watcher   *fsnotify.Watcher
	callbacks map[string][]func(Event)
	mu        sync.RWMutex
}

// Event represents a file system event
type Event struct {
	Path string `json:"path"`
	Type string `json:"type"` // create, modify, delete, rename
}

// NewFileWatcher creates a new FileWatcher
func NewFileWatcher() (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, utils.WrapError(1002, "无法创建文件监视器", err)
	}

	fw := &FileWatcher{
		watcher:   watcher,
		callbacks: make(map[string][]func(Event)),
	}

	// Start watching for events
	go fw.processEvents()

	return fw, nil
}

// Watch starts watching a path
func (fw *FileWatcher) Watch(path string, callback func(Event)) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Add path to watcher
	if err := fw.watcher.Add(path); err != nil {
		return utils.WrapError(1002, "无法监视路径", err)
	}

	// Register callback
	fw.callbacks[path] = append(fw.callbacks[path], callback)

	return nil
}

// Unwatch stops watching a path
func (fw *FileWatcher) Unwatch(path string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	fw.watcher.Remove(path)
	delete(fw.callbacks, path)
}

// Close closes the watcher
func (fw *FileWatcher) Close() {
	fw.watcher.Close()
}

// processEvents processes file system events
func (fw *FileWatcher) processEvents() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Determine event type
			var eventType string
			switch {
			case event.Op&fsnotify.Create == fsnotify.Create:
				eventType = "create"
			case event.Op&fsnotify.Write == fsnotify.Write:
				eventType = "modify"
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				eventType = "delete"
			case event.Op&fsnotify.Rename == fsnotify.Rename:
				eventType = "rename"
			default:
				continue
			}

			// Create event
			evt := Event{
				Path: event.Name,
				Type: eventType,
			}

			// Call callbacks
			fw.mu.RLock()
			for path, callbacks := range fw.callbacks {
				// Check if the event is for this path or a child
				if event.Name == path || isChildPath(path, event.Name) {
					for _, callback := range callbacks {
						go callback(evt)
					}
				}
			}
			fw.mu.RUnlock()

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			if utils.Log != nil {
				utils.Log.Error("文件监视器错误: %v", err)
			}
		}
	}
}

// isChildPath checks if child is a child path of parent
func isChildPath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != ".." && rel != "."
}

// WatchDirectory watches a directory recursively
func (fw *FileWatcher) WatchDirectory(path string, callback func(Event)) error {
	// Walk the directory and add all subdirectories
	return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Ignore errors
		}
		if info.IsDir() {
			if err := fw.Watch(walkPath, callback); err != nil {
				return err
			}
		}
		return nil
	})
}
