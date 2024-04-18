package main

import (
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/app/handler"
	"github.com/Tomas-vilte/GoMusicBot/internal/bot"
	"github.com/Tomas-vilte/GoMusicBot/internal/config"
	"log"
)

func main() {
	// Cargar la configuración desde el archivo .env
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Println("Error al cargar config:", err)
		return
	}

	// Crear el manejador de comandos
	commandHandler := handler.NewCommandHandler()
	registerCommands(commandHandler)

	// Crear y configurar el bot
	botDs, err := bot.NewBot(cfg, commandHandler.Handle)
	if err != nil {
		fmt.Println("Error en crear el bot: ", err)
		return
	}

	// Iniciar el bot
	if err := botDs.Run(); err != nil {
		log.Println("Error en correr el bot: ", err)
		return
	}
}

func registerCommands(handlers *handler.CommandHandler) {
	handlers.RegisterCommand("ping", &handler.PingCommand{})
	handlers.RegisterCommand("help", &handler.HelpCommand{})
	handlers.RegisterCommand("play", &handler.MusicCommand{})
}
