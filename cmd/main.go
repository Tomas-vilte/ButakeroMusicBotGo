package main

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/app/handler"
	"github.com/Tomas-vilte/GoMusicBot/internal/app/service"
	"github.com/Tomas-vilte/GoMusicBot/internal/config"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var sessionFactory discord.SessionFactory
	// Obtener la configuración del discord
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error al cargar la configuración del discord: %v", err)
		return
	}

	// Obtener la fábrica de sesiones de discord
	sessionFactory = &discord.ProductionBotSessionFactory{}

	// Obtener la sesión del discord
	session, err := sessionFactory.NewBotSession(cfg)
	if err != nil {
		log.Fatalf("Error al crear la sesión del discord: %v", err)
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
