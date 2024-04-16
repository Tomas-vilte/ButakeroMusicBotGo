package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	url2 "net/url"
	"time"
)

type SpotifyAuth struct {
	ClientID     string
	ClientSecret string
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func newSpotifyAuth(clientID, clientSecret string) *SpotifyAuth {
	return &SpotifyAuth{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}
func (s *SpotifyAuth) getAccessToken() (string, error) {
	url := "https://accounts.spotify.com/api/token"
	data := url2.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.ClientID+":"+s.ClientSecret)))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp AccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}
