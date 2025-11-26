package file

import (
	"errors"
	"os"
	"time"
)

// limitations : 20 MB per or 50 MB

var _ Watcher = (*IWatcher)(nil)

type Watcher interface {
	AddFile(path string) error
	AddDir(path string) error
	GetUpdatedFiles() ([]string, error)
}

type IWatcher struct {
	fileUpdates chan string
}

func NewWatcher() *IWatcher {
	return &IWatcher{
		fileUpdates: make(chan string),
	}
}

func (w *IWatcher) AddFile(path string) error {
	return w.watchFile(path)
}

func (w *IWatcher) AddDir(path string) error {
	return w.watchDir(path)
}

func (w *IWatcher) GetUpdatedFiles() ([]string, error) {
	files := make([]string, 1)
	files[0] = "README.md"

	return files, nil
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

func (w *IWatcher) watchDir(dirPath string) error {
	return errors.New("TODO")
}
