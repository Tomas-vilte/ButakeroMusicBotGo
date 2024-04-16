package bot

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

// Bot Representa el bot de ds
type Bot struct {
	session *discordgo.Session
}

type BotService interface {
	Open() error
	Close()
}

func NewBot(token string, handlers ...func(*discordgo.Session, *discordgo.MessageCreate)) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("Error al crear la session de discord:", err)
	}

	bot := &Bot{
		session: session,
	}

	for _, handler := range handlers {
		bot.session.AddHandler(handler)
	}
	return bot, nil
}

func (b *Bot) Open() error {
	return b.session.Open()
}

func (b *Bot) Close() {
	b.session.Close()
}
