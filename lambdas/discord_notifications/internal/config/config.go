package config

import "os"

type Config struct {
	WebhookURL       string
	QueueURL         string
	AwsRegion        string
	DiscordToken     string
	DiscordChannelID string
}

func LoadConfig() *Config {
	return &Config{
		WebhookURL:       os.Getenv("WEBHOOK_URL"),
		QueueURL:         os.Getenv("QUEUE_URL"),
		AwsRegion:        os.Getenv("AWS_REGION"),
		DiscordToken:     os.Getenv("DISCORD_TOKEN"),
		DiscordChannelID: os.Getenv("DISCORD_CHANNEL_ID"),
	}
}
