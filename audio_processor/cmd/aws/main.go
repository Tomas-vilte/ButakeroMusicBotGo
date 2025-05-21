package main

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/cmd/aws/server"
	"log"
)

func main() {
	if err := server.StartServer(); err != nil {
		log.Fatalf("Error al iniciar server: %v", err)
	}
}
