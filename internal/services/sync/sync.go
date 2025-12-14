package sync

import (
	"fmt"
	"path/filepath"

	"github.com/k0ff1l/tgcloudbot/internal/services/file"
	"github.com/k0ff1l/tgcloudbot/internal/services/telegram"
)

// SyncService handles file synchronization between local filesystem and Telegram
type SyncService struct {
	telegramBot telegram.Bot
	fileWatcher file.Watcher
	chatID      string
}

// NewSyncService creates a new sync service instance
func NewSyncService(bot telegram.Bot, watcher file.Watcher, chatID string) *SyncService {
	return &SyncService{
		telegramBot: bot,
		fileWatcher: watcher,
		chatID:      chatID,
	}
}

// SyncFile uploads a single file to Telegram
func (s *SyncService) SyncFile(filePath string) error {
	// Determine file type and send accordingly
	ext := filepath.Ext(filePath)

	caption := fmt.Sprintf("File: %s", filepath.Base(filePath))

	switch ext {
	case ".mp3", ".wav", ".ogg", ".m4a", ".flac":
		// Send as audio
		_, err := s.telegramBot.SendAudio(s.chatID, filePath, caption)
		if err != nil {
			return fmt.Errorf("failed to send audio: %w", err)
		}
	default:
		// Send as document
		_, err := s.telegramBot.SendDocument(s.chatID, filePath, caption)
		if err != nil {
			return fmt.Errorf("failed to send document: %w", err)
		}
	}

	return nil
}

// SyncDirectory watches a directory and syncs new/updated files
func (s *SyncService) SyncDirectory(dirPath string) error {
	// Add directory to watcher
	if err := s.fileWatcher.AddDir(dirPath); err != nil {
		return fmt.Errorf("failed to add directory to watcher: %w", err)
	}

	// Get updated files and sync them
	files, err := s.fileWatcher.GetUpdatedFiles()
	if err != nil {
		return fmt.Errorf("failed to get updated files: %w", err)
	}

	for _, filePath := range files {
		if err := s.SyncFile(filePath); err != nil {
			// Log error but continue with other files
			fmt.Printf("Error syncing file %s: %v\n", filePath, err)
			continue
		}
	}

	return nil
}

// StartContinuousSync starts continuous synchronization in background
func (s *SyncService) StartContinuousSync(dirPath string) error {
	// TODO: Implement continuous sync with goroutines
	// This would watch for file changes and automatically sync them
	return s.SyncDirectory(dirPath)
}
