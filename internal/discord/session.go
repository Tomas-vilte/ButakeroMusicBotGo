package discord

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/config"
	"github.com/bwmarrin/discordgo"
	"log"
)

var Session *discordgo.Session

func InitSession() (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + config.GetDiscordToken())
	if err != nil {
		log.Fatalf("Error en crear session con discord: %v", err)
		return nil, err
	}
	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentGuildMessageTyping | discordgo.IntentGuildVoiceStates | discordgo.IntentGuilds
	return session, nil
}

func OpenConnection(session *discordgo.Session) error {
	if err := session.Open(); err != nil {
		return err
	}
	return nil
}
