package spotify

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Device PlayRequest representa la estructura de la solicitud para iniciar/reanudar la reproducción
type Device struct {
	ID               string `json:"id"`
	IsActive         bool   `json:"is_active"`
	IsPrivateSession bool   `json:"is_private_session"`
	IsRestricted     bool   `json:"is_restricted"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	VolumePercent    int    `json:"volume_percent"`
}

const devicesURL = "https://api.spotify.com/v1/me/player/devices"

// DeviceService define la interfaz para obtener dispositivos de Spotify.
type DeviceService interface {
	GetDevices(accessToken string) ([]Device, error)
}

// SpotifyClient proporciona métodos para interactuar con los dispositivos de Spotify.
type SpotifyClient struct {
	HTTPClient *http.Client
}

// NewSpotifyClient crea una nueva instancia de SpotifyClient.
func NewSpotifyClient(client *http.Client) *SpotifyClient {
	return &SpotifyClient{
		HTTPClient: client,
	}
}

// GetDevices obtiene la lista de dispositivos disponibles para un usuario.
func (s *SpotifyClient) GetDevices(accessToken string) ([]Device, error) {
	log.Println("Obteniendo dispositivos de Spotify...")

	req, err := http.NewRequest("GET", devicesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud HTTP: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error haciendo solicitud HTTP: %v", err)
	}
	defer resp.Body.Close()
	var devices struct {
		Devices []Device `json:"devices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, fmt.Errorf("error decodificando respuesta JSON: %v", err)
	}

	log.Println("Dispositivos obtenidos exitosamente")
	return devices.Devices, nil
}

// Printer define la interfaz para imprimir dispositivos.
type Printer interface {
	Print(device []Device)
}

type ConsolePrinter struct{}

// Print imprime dispositivos en la consola.
func (cp *ConsolePrinter) Print(devices []Device) {
	log.Println("Imprimiendo dispositivos:")
	for _, device := range devices {
		log.Printf("ID: %s, Name: %s, Type: %s, Is Active: %t, Volume Percent: %d\n", device.ID, device.Name, device.Type, device.IsActive, device.VolumePercent)
	}
}
