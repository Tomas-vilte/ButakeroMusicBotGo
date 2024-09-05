package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type (
	// YouTubeService define la interfaz para interactuar con el servicio de YouTube.
	YouTubeService interface {
		// GetVideoDetails obtiene los detalles del video usando su ID.
		GetVideoDetails(ctx context.Context, videoID string) (*VideoDetails, error)
		// SearchVideoID busca el ID del primer video que coincida con la consulta dada.
		SearchVideoID(ctx context.Context, input string) (string, error)
	}

	// VideoDetails contiene los detalles de un video de YouTube.
	VideoDetails struct {
		Title       string    // Título del video.
		Description string    // Descripción del video.
		ChannelName string    // Nombre del canal que subió el video.
		Duration    string    // Duración del video en formato ISO 8601.
		PublishedAt time.Time // Fecha de publicación del video.
	}

	// YouTubeClient es un cliente para interactuar con la API de YouTube.
	YouTubeClient struct {
		ApiKey     string       // Clave API para autenticar las solicitudes.
		BaseURL    string       // URL base para las solicitudes a la API de YouTube.
		HttpClient *http.Client // Cliente HTTP para hacer las solicitudes.
	}
)

// NewYouTubeClient crea una nueva instancia de YouTubeClient.
func NewYouTubeClient(apiKey string) *YouTubeClient {
	return &YouTubeClient{
		ApiKey:  apiKey,
		BaseURL: "https://youtube.googleapis.com/youtube/v3",
		HttpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetVideoDetails obtiene los detalles del video usando su ID.
func (c *YouTubeClient) GetVideoDetails(ctx context.Context, videoID string) (*VideoDetails, error) {
	url := fmt.Sprintf("%s/videos?part=snippet,contentDetails&id=%s&key=%s", c.BaseURL, videoID, c.ApiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error al crear la solicitud: %w", err)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al hacer la solicitud a la API de YouTube: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error en la API de YouTube, código de estado: %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			Snippet struct {
				Title        string `json:"title"`
				Description  string `json:"description"`
				ChannelTitle string `json:"channelTitle"`
				PublishedAt  string `json:"publishedAt"`
			} `json:"snippet"`
			ContentDetails struct {
				Duration string `json:"duration"`
			} `json:"contentDetails"`
		} `json:"items"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("error al decodificar la respuesta de la API de YouTube: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no se encontró el video con el ID proporcionado")
	}

	item := result.Items[0]
	publishedAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)

	videoDetails := &VideoDetails{
		Title:       item.Snippet.Title,
		Description: item.Snippet.Description,
		ChannelName: item.Snippet.ChannelTitle,
		Duration:    item.ContentDetails.Duration,
		PublishedAt: publishedAt,
	}

	return videoDetails, nil
}

// SearchVideoID busca el ID del video basado en la entrada proporcionada.
func (c *YouTubeClient) SearchVideoID(ctx context.Context, input string) (string, error) {

	// Verifica si la entrada es una URL completa
	if strings.Contains(input, "youtube.com/watch") || strings.Contains(input, "youtu.be/") {
		videoID, err := ExtractVideoIDFromURL(input)
		if err != nil {
			return "", err
		}
		return videoID, nil
	}
	encodedQuery := url.QueryEscape(input)
	endpoint := fmt.Sprintf("%s/search?part=id&q=%s&key=%s&type=video&maxResults=1", c.BaseURL, encodedQuery, c.ApiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("error al crear la solicitud: %w", err)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error al hacer la solicitud a la API de YouTube: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error en la API de YouTube, código de estado: %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error al decodificar la respuesta de la API de YouTube: %w", err)
	}

	if len(result.Items) == 0 {
		return "", fmt.Errorf("no se encontraron videos para la consulta: %s", input)
	}

	return result.Items[0].ID.VideoID, nil
}

// ExtractVideoIDFromURL extrae el ID del video de una URL de YouTube.
func ExtractVideoIDFromURL(videoURL string) (string, error) {
	// Expresión regular para extraer el ID del video de una URL de YouTube
	re := regexp.MustCompile(`(?:https?://)?(?:www\.)?youtube\.com/(?:watch\?v=|embed/|v/|.+/v/|.+/embed/|user/(?:\w+/)?\w+/\w+/|watch\?.*v=|shorts/|playlist\?list=)([\w-]{11})`)
	matches := re.FindStringSubmatch(videoURL)
	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", fmt.Errorf("URL de YouTube invalida: %s", videoURL)
}
