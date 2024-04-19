package main

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/app/handler"
	"github.com/Tomas-vilte/GoMusicBot/internal/app/service"
	"github.com/Tomas-vilte/GoMusicBot/internal/bot"
	"github.com/Tomas-vilte/GoMusicBot/internal/config"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var sessionFactory bot.SessionFactory
	// Obtener la configuración del bot
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error al cargar la configuración del bot: %v", err)
		return
	}

	// Obtener la fábrica de sesiones de bot
	sessionFactory = &bot.ProductionBotSessionFactory{}

	// Obtener la sesión del bot
	session, err := sessionFactory.NewBotSession(cfg)
	if err != nil {
		log.Fatalf("Error al crear la sesión del bot: %v", err)
		return
	}

	// Crear el servicio de ping
	pingService := &service.PingServiceImpl{}

	// Crear el manejador de comandos de ping
	pingCommandHandler := &handler.PingCommandHandler{
		PingService: pingService,
	}

	// Registrar y manejar comandos
	_, err = pingCommandHandler.RegisterCommands(session)
	if err != nil {
		log.Fatalf("Error al registrar comandos: %v", err)
		return
	}

	session.AddHandler(pingCommandHandler.HandleInteraction)

	log.Println("Bot is now running. Press CTRL-C to exit.")

	// Esperar a que se reciba una señal de cierre (CTRL-C)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cerrar la sesión de Discord
	session.Close()
}
