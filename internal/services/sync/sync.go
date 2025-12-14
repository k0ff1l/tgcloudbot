package sync

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/k0ff1l/tgcloudbot/internal/services/file"
	"github.com/k0ff1l/tgcloudbot/internal/services/telegram"
)

// SyncService handles file synchronization between local filesystem and Telegram
type SyncService struct {
	bot         telegram.Client
	fileWatcher file.Watcher
	chatID      string
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewSyncService creates a new sync service instance
func NewSyncService(bot telegram.Client, watcher file.Watcher, chatID string) *SyncService {
	ctx, cancel := context.WithCancel(context.Background())
	return &SyncService{
		bot:         bot,
		fileWatcher: watcher,
		chatID:      chatID,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// SyncFile uploads a single file to Telegram
func (s *SyncService) SyncFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("filePath cannot be empty")
	}

	// Determine file type and send accordingly
	ext := strings.ToLower(filepath.Ext(filePath))

	// TODO: delete - remove caption generation
	caption := fmt.Sprintf("File: %s", filepath.Base(filePath))

	switch ext {
	case ".mp3", ".wav", ".ogg", ".m4a", ".flac", ".aac", ".opus":
		// Send as audio
		_, err := s.bot.SendAudio(s.chatID, filePath, caption)
		if err != nil {
			return fmt.Errorf("failed to send audio: %w", err)
		}
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
		// Send as photo
		_, err := s.bot.SendPhoto(s.chatID, filePath, caption)
		if err != nil {
			return fmt.Errorf("failed to send photo: %w", err)
		}
	case ".mp4", ".avi", ".mov", ".mkv", ".webm", ".flv", ".wmv":
		// Send as video
		_, err := s.bot.SendVideo(s.chatID, filePath, caption)
		if err != nil {
			return fmt.Errorf("failed to send video: %w", err)
		}
	default:
		// Send as document
		_, err := s.bot.SendDocument(s.chatID, filePath, caption)
		if err != nil {
			return fmt.Errorf("failed to send document: %w", err)
		}
	}

	return nil
}

// SyncDirectory watches a directory and syncs new/updated files
func (s *SyncService) SyncDirectory(dirPath string) error {
	if dirPath == "" {
		return fmt.Errorf("dirPath cannot be empty")
	}

	// Add directory to watcher
	if err := s.fileWatcher.AddDir(dirPath); err != nil {
		return fmt.Errorf("failed to add directory to watcher: %w", err)
	}

	// Get updated files and sync them
	files, err := s.fileWatcher.GetUpdatedFiles()
	if err != nil {
		return fmt.Errorf("failed to get updated files: %w", err)
	}

	if files == nil {
		return fmt.Errorf("fileWatcher returned nil files list")
	}

	for _, filePath := range files {
		if filePath == "" {
			continue // Skip empty paths
		}

		if err := s.SyncFile(filePath); err != nil {
			// Log error but continue with other files
			fmt.Printf("Error syncing file %s: %v\n", filePath, err)
			continue
		}
	}

	return nil
}

// StartContinuousSync starts continuous synchronization in background
// It periodically checks for file changes and syncs them
func (s *SyncService) StartContinuousSync(dirPath string, interval time.Duration) error {
	if dirPath == "" {
		return fmt.Errorf("dirPath cannot be empty")
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}

	// Add directory to watcher
	if err := s.fileWatcher.AddDir(dirPath); err != nil {
		return fmt.Errorf("failed to add directory to watcher: %w", err)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Initial sync
		if err := s.syncDirectoryOnce(); err != nil {
			fmt.Printf("Error in initial sync for %s: %v\n", dirPath, err)
		}

		for {
			select {
			case <-s.ctx.Done():
				fmt.Printf("Stopping continuous sync for %s\n", dirPath)
				return
			case <-ticker.C:
				if err := s.syncDirectoryOnce(); err != nil {
					fmt.Printf("Error syncing directory %s: %v\n", dirPath, err)
				}
			}
		}
	}()

	return nil
}

// syncDirectoryOnce performs a single sync operation for a directory
func (s *SyncService) syncDirectoryOnce() error {
	files, err := s.fileWatcher.GetUpdatedFiles()
	if err != nil {
		return fmt.Errorf("failed to get updated files: %w", err)
	}

	if files == nil {
		return nil // No files to sync
	}

	for _, filePath := range files {
		if filePath == "" {
			continue // Skip empty paths
		}

		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		default:
		}

		if err := s.SyncFile(filePath); err != nil {
			// Log error but continue with other files
			fmt.Printf("Error syncing file %s: %v\n", filePath, err)
			continue
		}
	}

	return nil
}

// Stop stops all continuous sync operations
func (s *SyncService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
		s.wg.Wait()
	}
}
