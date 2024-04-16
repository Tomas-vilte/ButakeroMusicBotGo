package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	DiscordBotToken string
}

// NewConfig crea una nueva instancia de Config cargando variables de entorno.
func NewConfig() (*Config, error) {
	// Carga las variables de entorno desde el archivo .env
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	// Lee las variables de entorno
	discordBotToken := os.Getenv("DISCORD_BOT_TOKEN")

	// Crea una instancia de Config con las variables cargadas
	cfg := &Config{DiscordBotToken: discordBotToken}
	return cfg, nil
}
