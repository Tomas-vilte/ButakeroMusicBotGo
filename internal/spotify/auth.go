package spotify

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	url2 "net/url"
	"strings"
	"time"
)

// AccessTokenResponse representa la respuesta de un token de acceso.
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// AuthClient define la interfaz para la autenticación con Spotify.
type AuthClient interface {
	GetAccessToken() (string, error)
}

// SpotifyAuth representa la autenticación de Spotify.
type SpotifyAuth struct {
	ClientID     string
	ClientSecret string
}

// NewSpotifyAuth crea una nueva instancia de SpotifyAuth.
func NewSpotifyAuth(clientID, clientSecret string) *SpotifyAuth {
	return &SpotifyAuth{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

// GetAccessToken obtiene un token de acceso de Spotify.
func (s *SpotifyAuth) GetAccessToken() (string, error) {
	url := "https://accounts.spotify.com/api/token"
	data := url2.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %v", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.ClientID+":"+s.ClientSecret)))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error haciendo solicitud HTTP: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp AccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		log.Printf("Error decodificando respuesta: %v", err)
		return "", err
	}

	return tokenResp.AccessToken, nil
}
