package discord

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/config"
	"github.com/bwmarrin/discordgo"
	"log"
)

// SessionFactory define la interfaz para la fábrica de sesiones de discord.
type SessionFactory interface {
	NewBotSession(cfg *config.Config) (*discordgo.Session, error)
}

// ProductionBotSessionFactory es una implementación concreta de SessionFactory para producción.
type ProductionBotSessionFactory struct{}

// NewBotSession crea una nueva sesión de discord y la devuelve.
func (f *ProductionBotSessionFactory) NewBotSession(cfg *config.Config) (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + cfg.DiscordBotToken)
	if err != nil {
		log.Fatalf("Error al crear la sesion de discord: %v", err)
		return nil, err
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentGuildMessageTyping | discordgo.IntentGuildVoiceStates | discordgo.IntentGuilds

	err = session.Open()
	if err != nil {
		log.Fatalf("Error al estar la sesion de discord: %v", err)
		return nil, err
	}

	log.Println("Bot corriendo, apreta CTRL-C para salir")
	return session, nil
}
