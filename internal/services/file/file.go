package file

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TODO: implement filter names (.txt, .swp, ...) from config.go regexp

// limitations : 20 MB per or 50 MB

var _ Watcher = (*IWatcher)(nil)

type Watcher interface {
	AddFile(path string) error
	AddDir(path string) error
	GetUpdatedFiles() ([]string, error)
}

type watchedFile struct {
	modTime  time.Time
	lastSync time.Time
	path     string
	size     int64
}

type IWatcher struct {
	fileUpdates  chan string
	watchedDirs  map[string]bool
	watchedFiles map[string]*watchedFile
	mu           sync.RWMutex
}

func NewWatcher() *IWatcher {
	return &IWatcher{
		fileUpdates:  make(chan string, 100),
		watchedDirs:  make(map[string]bool),
		watchedFiles: make(map[string]*watchedFile),
	}
}

func (w *IWatcher) AddFile(path string) error {
	return w.watchFile(path)
}

func (w *IWatcher) AddDir(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.watchedDirs[path] {
		return nil // Already watching
	}

	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	w.watchedDirs[path] = true

	return nil
}

func (w *IWatcher) GetUpdatedFiles() ([]string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	var updatedFiles []string

	// Check all watched directories
	for dirPath := range w.watchedDirs {
		files, err := w.scanDirectory(dirPath)
		if err != nil {
			continue // Skip directories that can't be scanned
		}

		for _, filePath := range files {
			info, err := os.Stat(filePath)
			if err != nil {
				continue // Skip files that can't be stat'd
			}

			// Check if file is new or modified
			watched, exists := w.watchedFiles[filePath]
			if !exists {
				// New file
				w.watchedFiles[filePath] = &watchedFile{
					path:     filePath,
					modTime:  info.ModTime(),
					size:     info.Size(),
					lastSync: time.Now(),
				}
				updatedFiles = append(updatedFiles, filePath)
			} else if info.ModTime().After(watched.modTime) || info.Size() != watched.size {
				// File was modified
				watched.modTime = info.ModTime()
				watched.size = info.Size()
				watched.lastSync = time.Now()

				updatedFiles = append(updatedFiles, filePath)
			}
		}
	}

	return updatedFiles, nil
}

// scanDirectory recursively scans a directory and returns all file paths
func (w *IWatcher) scanDirectory(dirPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files/dirs that can't be accessed
		}

		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func (w *IWatcher) watchFile(filePath string) error {
	initialStat, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// TODO: go func () ... for {-> channel}
	for {
		stat, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}
