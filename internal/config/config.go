package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	DiscordBotToken string
	ChannelID       string
	ServerID        string
	BotRoleID       string
	BotID           string
}

// NewConfig crea una nueva instancia de Config cargando variables de entorno.
func NewConfig() (*Config, error) {
	// Carga las variables de entorno desde el archivo .env
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	// Lee las variables de entorno
	discordBotToken := os.Getenv("DISCORD_BOT_TOKEN")
	channelID := os.Getenv("CHANNEL_ID")
	serverID := os.Getenv("SERVER_ID")
	botRoleID := os.Getenv("BOT_ROLE_ID")
	botID := os.Getenv("BOT_ID")

	// Crea una instancia de Config con las variables cargadas
	cfg := &Config{
		DiscordBotToken: discordBotToken,
		ChannelID:       channelID,
		ServerID:        serverID,
		BotRoleID:       botRoleID,
		BotID:           botID,
	}
	return cfg, nil
}
