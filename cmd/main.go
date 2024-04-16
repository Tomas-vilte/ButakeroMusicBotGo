package main

import (
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/app/handler"
	"github.com/Tomas-vilte/GoMusicBot/internal/bot"
	"github.com/Tomas-vilte/GoMusicBot/internal/config"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Cargar la configuración desde el archivo .env
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Println("Error al cargar config:", err)
		return
	}

	commandHandler := handler.NewCommandHandler()
	commandHandler.RegisterCommand("ping", &handler.PingCommand{})
	commandHandler.RegisterCommand("help", &handler.HelpCommand{})

	botDs, err := bot.NewBot(cfg.DiscordBotToken, commandHandler.Handle)
	if err != nil {
		fmt.Println("Error creating bot: ", err)
		return
	}

	// Abrir sesión de Discord
	err = botDs.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")

	// Esperar a que se reciba una señal de cierre (CTRL-C)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cerrar sesión de Discord
	botDs.Close()
}
