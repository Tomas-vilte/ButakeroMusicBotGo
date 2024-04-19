package bot

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/config"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Bot Representa el bot de ds
type Bot struct {
	session *discordgo.Session
}

func NewBot(config *config.Config, handlers ...func(*discordgo.Session, *discordgo.MessageCreate)) (*Bot, error) {
	session, err := discordgo.New("Bot " + config.DiscordBotToken)
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

// Run inicia el bot y comienza a escuchar los eventos.
func (b *Bot) Run() error {
	err := b.session.Open()
	if err != nil {
		log.Fatalf("Error al abrir session con discord: %v", err)
		return err
	}
	defer b.session.Close()

	log.Println("Bot corriendo, Apreta CTRL + C para cerrarlo")

	// Esperar a que se reciba una señal de cierre (CTRL-C)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	return nil
}
