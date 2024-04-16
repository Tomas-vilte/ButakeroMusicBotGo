package main

import (
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/spotify"
	"log"
	"net/http"
)

func main() {
	clientHTTP := &http.Client{}

	// Crear una instancia de SpotifyClient
	clientSpotify := spotify.NewSpotifyClient(clientHTTP)

	auth := spotify.NewSpotifyAuth("94739f386d624bc89b3a6f95e57a4b88", "a4ca184ff4914bd6811b7816bb8d32e2")
	accessToken, err := auth.GetAccessToken()
	if err != nil {
		log.Fatalf("Error al obtener token de acceso: %v", err)
	}

	devices, err := clientSpotify.GetDevices(accessToken)
	if err != nil {
		log.Fatalf("Error al obtener la devices: %v", err)
	}

	for _, device := range devices {
		fmt.Println(device.Name)
	}
}
