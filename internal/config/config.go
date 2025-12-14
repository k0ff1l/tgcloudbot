package config

import (
	"fmt"
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
	// TODO: regexp whitelist/blacklist
}

const (
	defaultAPIURL   = "https://api.telegram.org/bot"
	defaultInterval = 1 * time.Second
)

func New() *Config {
	cfg := &Config{
		BotToken:     mustGetEnv("TELEGRAM_BOT_TOKEN"),
		ChatID:       mustGetEnv("TELEGRAM_CHAT_ID"),
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

	return cfg
}

func mustGetEnv(key string) string {
	var val string

	if val = os.Getenv(key); val == "" {
		panic(fmt.Sprintf("environment variable `%s` not set", key))
	}

	return val
}
