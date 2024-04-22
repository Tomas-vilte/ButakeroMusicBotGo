package config

import (
	"github.com/joho/godotenv"
	"os"
)

var config *Config

type Config struct {
	DiscordBotToken     string
	BotStatus           string
	BotGuildJoinMessage string
	ChannelID           string
	ServerID            string
	BotRoleID           string
	BotID               string
	BotPrefix           string
}

// NewConfig crea una nueva instancia de Config cargando variables de entorno.
func NewConfig() (*Config, error) {
	// Carga las variables de entorno desde el archivo .env
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	// Lee las variables de entorno
	discordBotToken := os.Getenv("DISCORD_BOT_TOKEN")
	botStatus := os.Getenv("BOT_STATUS")
	botGuildJoinMessage := os.Getenv("BOT_GUILD_JOIN_MESSAGE")
	channelID := os.Getenv("CHANNEL_ID")
	serverID := os.Getenv("SERVER_ID")
	botRoleID := os.Getenv("BOT_ROLE_ID")
	botID := os.Getenv("BOT_ID")
	botPrefix := os.Getenv("BOT_PREFIX")

	// Crea una instancia de Config con las variables cargadas
	cfg := &Config{
		DiscordBotToken:     discordBotToken,
		BotStatus:           botStatus,
		BotGuildJoinMessage: botGuildJoinMessage,
		ChannelID:           channelID,
		ServerID:            serverID,
		BotRoleID:           botRoleID,
		BotID:               botID,
		BotPrefix:           botPrefix,
	}
	return cfg, nil
}

func GetDiscordToken() string {
	return config.DiscordBotToken
}

func GetChannelID() string {
	return config.ChannelID
}

func GetServerID() string {
	return config.ServerID
}

func GetBotRoleID() string {
	return config.BotRoleID
}

func GetBotID() string {
	return config.BotID
}

func GetBotStatus() string {
	return config.BotStatus
}

func GetBotGuildJoinMessage() string {
	return config.BotGuildJoinMessage
}

func GetBotPrefix() string {
	return config.BotPrefix
}
