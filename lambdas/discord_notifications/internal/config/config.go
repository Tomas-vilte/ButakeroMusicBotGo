package config

import (
	"os"
)

type Config struct {
	DiscordToken string
}

func LoadConfig() *Config {
	config := &Config{
		DiscordToken: os.Getenv("DISCORD_TOKEN"),
	}
	return config
}
