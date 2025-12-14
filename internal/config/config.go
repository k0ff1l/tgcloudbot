package config

import (
	"os"
)

type Config struct {
	BotToken string
	ChatID   string
	APIURL   string
}

const (
	defaultAPIURL = "https://api.telegram.org/bot"
)

func New() *Config {
	cfg := &Config{
		BotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		ChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
		APIURL:   defaultAPIURL,
	}

	// TODO: YAML config parse
	// TODO: godotenv parse for credentials (sensitive information)

	return cfg
}
