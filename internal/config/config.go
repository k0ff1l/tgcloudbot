package config

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

type Config struct {
	BotToken        string
	ChatID          string
	APIURL          string
	WatchDirs       []string
	WhitelistRegexp []*regexp.Regexp
	BlacklistRegexp []*regexp.Regexp
	SyncInterval    time.Duration
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
	cfg.WatchDirs = getStringsSlice(os.Getenv("TELEGRAM_WATCH_DIRS"))
	whitelist := getStringsSlice(os.Getenv("WHITELIST_REGEXP"))

	cfg.WhitelistRegexp = make([]*regexp.Regexp, 0, len(whitelist))

	for _, regString := range whitelist {
		regex, err := regexp.Compile(regString)
		if err != nil {
			log.Printf("Failed to compile regexp '%s': %v", regString, err)
		}

		cfg.WhitelistRegexp = append(cfg.WhitelistRegexp, regex)
	}

	blacklist := getStringsSlice(os.Getenv("BLACKLIST_REGEXP"))

	cfg.BlacklistRegexp = make([]*regexp.Regexp, 0, len(blacklist))

	for _, regString := range blacklist {
		regex, err := regexp.Compile(regString)
		if err != nil {
			log.Printf("Failed to compile regexp '%s': %v", regString, err)
		}

		cfg.BlacklistRegexp = append(cfg.BlacklistRegexp, regex)
	}

	return cfg
}

func getStringsSlice(val string) []string {
	slice := strings.Split(val, ",")

	result := make([]string, 0, len(slice))

	for _, str := range slice {
		str = strings.TrimSpace(str)
		if str != "" {
			result = append(result, str)
		}
	}

	return result
}

func mustGetEnv(key string) string {
	var val string

	if val = os.Getenv(key); val == "" {
		panic(fmt.Sprintf("environment variable `%s` not set", key))
	}

	return val
}
