package config

type Config struct {
	BotToken string
}

func New() *Config {
	// YAML config parse

	// godotenv parse for credentials (sensitive information)

	return &Config{}
}
