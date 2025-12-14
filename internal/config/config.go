package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	BotToken     string
	ChatID       string
	APIURL       string
	WatchDirs    []string
	SyncInterval time.Duration
}

const (
	defaultAPIURL   = "https://api.telegram.org/bot"
	defaultInterval = 5 * time.Second
)

func New() *Config {
	cfg := &Config{
		BotToken:     os.Getenv("TELEGRAM_BOT_TOKEN"),
		ChatID:       os.Getenv("TELEGRAM_CHAT_ID"),
		APIURL:       defaultAPIURL,
		SyncInterval: defaultInterval,
	}

	// Parse watch directories from environment variable (comma-separated)
	watchDirsEnv := os.Getenv("TELEGRAM_WATCH_DIRS")
	if watchDirsEnv != "" {
		dirs := strings.Split(watchDirsEnv, ",")
		cfg.WatchDirs = make([]string, 0, len(dirs))
		for _, dir := range dirs {
			trimmed := strings.TrimSpace(dir)
			if trimmed != "" {
				cfg.WatchDirs = append(cfg.WatchDirs, trimmed)
			}
		}
	}

	// Parse sync interval from environment variable (in seconds)
	if intervalStr := os.Getenv("TELEGRAM_SYNC_INTERVAL"); intervalStr != "" {
		if interval, err := time.ParseDuration(intervalStr); err == nil {
			cfg.SyncInterval = interval
		}
	}

	// TODO: YAML config parse
	// TODO: godotenv parse for credentials (sensitive information)

	return cfg
}
