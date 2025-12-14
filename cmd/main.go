package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/k0ff1l/tgcloudbot/internal/config"
	"github.com/k0ff1l/tgcloudbot/internal/services/file"
	"github.com/k0ff1l/tgcloudbot/internal/services/sync"
	"github.com/k0ff1l/tgcloudbot/internal/services/telegram"
)

func main() {
	cfg := config.New()

	bot := telegram.NewBot(cfg.APIURL, cfg.BotToken)
	watcher := file.NewWatcher()
	syncService := sync.NewSyncService(bot, watcher, cfg.ChatID)

	// Send startup notification
	msg, err := bot.SendMessage(cfg.ChatID, "Bot started successfully! Starting file synchronization...")
	if err != nil {
		log.Printf("Failed to send startup message: %v", err)
	} else {
		log.Printf("Startup message sent successfully! Message ID: %d\n", msg.MessageID)
	}

	// Start continuous sync for configured directories
	if len(cfg.WatchDirs) > 0 {
		log.Printf("Starting continuous sync for %d directory(ies)...\n", len(cfg.WatchDirs))

		for _, dirPath := range cfg.WatchDirs {
			if err := syncService.StartContinuousSync(dirPath, cfg.SyncInterval); err != nil {
				log.Printf("Failed to start continuous sync for %s: %v\n", dirPath, err)
			} else {
				log.Printf("Started continuous sync for: %s (interval: %v)\n", dirPath, cfg.SyncInterval)
			}
		}
	} else {
		log.Println("No watch directories configured. Set TELEGRAM_WATCH_DIRS environment variable.")
		log.Println("Example: TELEGRAM_WATCH_DIRS=/path/to/dir1,/path/to/dir2")
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("Bot is running. Press Ctrl+C to stop...")

	// Wait for interrupt signal
	<-sigChan
	log.Println("\nShutting down gracefully...")

	syncService.Stop()

	if _, err := bot.SendMessage(cfg.ChatID, "Bot is shutting down. Goodbye!"); err != nil {
		log.Printf("Failed to send shutdown message: %v", err)
	}

	log.Println("Bot stopped successfully.")
}
