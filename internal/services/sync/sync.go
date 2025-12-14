package sync

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/k0ff1l/tgcloudbot/internal/services/file"
	"github.com/k0ff1l/tgcloudbot/internal/services/telegram"
)

// Service handles file synchronization between local filesystem and Telegram
type Service struct {
	bot         telegram.Client
	fileWatcher file.Watcher
	ctx         context.Context
	cancel      context.CancelFunc
	chatID      string
	wg          sync.WaitGroup
	mu          sync.RWMutex
}

// NewSyncService creates a new sync service instance
func NewSyncService(bot telegram.Client, fw file.Watcher, chatID string) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		bot:         bot,
		fileWatcher: fw,
		chatID:      chatID,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// SyncFile uploads a single file to Telegram
func (s *Service) SyncFile(filePath string) error {
	if filePath == "" {
		return errors.New("filePath cannot be empty")
	}

	// Determine file type and send accordingly
	ext := strings.ToLower(filepath.Ext(filePath))

	// TODO: delete - remove caption generation
	caption := "File: " + filepath.Base(filePath)

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
func (s *Service) SyncDirectory(dirPath string) error {
	if dirPath == "" {
		return errors.New("dirPath cannot be empty")
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
		return errors.New("fileWatcher returned nil files list")
	}

	for _, filePath := range files {
		if filePath == "" {
			continue // Skip empty paths
		}

		if err := s.SyncFile(filePath); err != nil {
			// Log error but continue with other files
			log.Printf("Error syncing file %s: %v\n", filePath, err)
			continue
		}
	}

	return nil
}

// StartContinuousSync starts continuous synchronization in background
// It periodically checks for file changes and syncs them
func (s *Service) StartContinuousSync(dirPath string, interval time.Duration) error {
	if dirPath == "" {
		return errors.New("dirPath cannot be empty")
	}

	if interval <= 0 {
		interval = 5 * time.Second
	}

	// Add directory to watcher
	if err := s.fileWatcher.AddDir(dirPath); err != nil {
		return fmt.Errorf("failed to add directory to watcher: %w", err)
	}

	s.wg.Go(func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Initial sync
		if err := s.syncDirectoryOnce(); err != nil {
			log.Printf("Error in initial sync for %s: %v\n", dirPath, err)
		}

		for {
			select {
			case <-s.ctx.Done():
				log.Printf("Stopping continuous sync for %s\n", dirPath)
				return
			case <-ticker.C:
				if err := s.syncDirectoryOnce(); err != nil {
					log.Printf("Error syncing directory %s: %v\n", dirPath, err)
				}
			}
		}
	})

	return nil
}

// Stop stops all continuous sync operations
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
		s.wg.Wait()
	}
}

// syncDirectoryOnce performs a single sync operation for a directory
func (s *Service) syncDirectoryOnce() error {
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
			log.Printf("Error syncing file %s: %v\n", filePath, err)
			continue
		}
	}

	return nil
}
