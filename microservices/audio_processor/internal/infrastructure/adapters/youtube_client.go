package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// YouTubeClient es un cliente para interactuar con la API de YouTube.
type YouTubeClient struct {
	ApiKey     string       // Clave API para autenticar las solicitudes.
	BaseURL    string       // URL base para las solicitudes a la API de YouTube.
	HttpClient *http.Client // Cliente HTTP para hacer las solicitudes.
	log        logger.Logger
}

// NewYouTubeClient crea una nueva instancia de YouTubeClient.
func NewYouTubeClient(apiKey string, log logger.Logger) *YouTubeClient {
	return &YouTubeClient{
		ApiKey:  apiKey,
		BaseURL: "https://youtube.googleapis.com/youtube/v3",
		HttpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log,
	}
}

// GetVideoDetails obtiene los detalles del video usando su ID.
func (c *YouTubeClient) GetVideoDetails(ctx context.Context, videoID string) (*model.MediaDetails, error) {
	c.log.With(zap.String("component", "YouTubeClient"), zap.String("video_id", videoID), zap.String("method", "GetVideoDetails"))
	c.log.Info("Obteniendo detalles del video")

	endpoint := fmt.Sprintf("%s/videos?part=snippet,contentDetails&id=%s&key=%s", c.BaseURL, videoID, c.ApiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		c.log.Error("Error al crear la solicitud HTTP", zap.Error(err))
		return nil, fmt.Errorf("error al crear la solicitud: %w", err)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		c.log.Error("Error al ejecutar la solicitud HTTP", zap.Error(err), zap.String("endpoint", endpoint))
		return nil, fmt.Errorf("error al hacer la solicitud a la API de YouTube: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Error("Error al cerrar el body de la respuesta", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.log.Error("Error en la API de YouTube", zap.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("API respondió con código %d", resp.StatusCode)
	}

	urlVideo := fmt.Sprintf("https://youtube.com/watch?v=%s", videoID)
	var result struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Thumbnails struct {
					Default struct {
						URL string `json:"url"`
					} `json:"default"`
				} `json:"thumbnails"`
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
		c.log.Error("Error al decodificar la respuesta de la API de YouTube", zap.Error(err))
		return nil, fmt.Errorf("error al decodificar la respuesta de la API de YouTube: %w", err)
	}

	if len(result.Items) == 0 {
		c.log.Warn("No se encontró el video con el ID proporcionado")
		return nil, fmt.Errorf("no se encontró el video con el ID proporcionado")
	}

	item := result.Items[0]
	publishedAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)

	videoDetails := &model.MediaDetails{
		Title:       item.Snippet.Title,
		ID:          item.ID,
		Description: item.Snippet.Description,
		Creator:     item.Snippet.ChannelTitle,
		Duration:    item.ContentDetails.Duration,
		Thumbnail:   item.Snippet.Thumbnails.Default.URL,
		PublishedAt: publishedAt,
		URL:         urlVideo,
	}

	return videoDetails, nil
}

// SearchVideoID busca el ID del video basado en la entrada proporcionada.
func (c *YouTubeClient) SearchVideoID(ctx context.Context, input string) (string, error) {
	c.log.With(zap.String("component", "YouTubeClient"), zap.String("input", input), zap.String("method", "SearchVideoID"))
	c.log.Info("Buscando ID del video")

	// Verifica si la entrada es una URL completa
	if strings.Contains(input, "youtube.com/watch") || strings.Contains(input, "youtu.be/") {
		videoID, err := ExtractVideoIDFromURL(input)
		if err != nil {
			c.log.Error("Error al extraer ID del video de la URL", zap.Error(err), zap.String("url", input))
			return "", err
		}
		return videoID, nil
	}
	encodedQuery := url.QueryEscape(input)
	endpoint := fmt.Sprintf("%s/search?part=id&q=%s&key=%s&type=video&maxResults=1", c.BaseURL, encodedQuery, c.ApiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		c.log.Error("Error al crear la solicitud HTTP", zap.Error(err))
		return "", fmt.Errorf("error al crear la solicitud: %w", err)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		c.log.Error("Error al ejecutar la solicitud HTTP", zap.Error(err))
		return "", fmt.Errorf("error al hacer la solicitud a la API de YouTube: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Error("Error al cerrar el body de la respuesta", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.log.Error("Error en la API de YouTube, código de estado", zap.Int("status_code", resp.StatusCode))
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
		c.log.Error("Error al decodificar la respuesta de la API de YouTube", zap.Error(err))
		return "", fmt.Errorf("error al decodificar la respuesta de la API de YouTube: %w", err)
	}

	if len(result.Items) == 0 {
		c.log.Warn("No se encontraron videos para la consulta", zap.String("input", input))
		return "", fmt.Errorf("no se encontraron videos para la consulta: %s", input)
	}

	return result.Items[0].ID.VideoID, nil
}

// ExtractVideoIDFromURL extrae el ID del video de una URL de YouTube.
func ExtractVideoIDFromURL(videoURL string) (string, error) {
	re := regexp.MustCompile(`^(?:https?://)?(?:www\.)?(?:youtube\.com/(?:watch\?v=|embed/|v/|.+/(?:embed|v)/|shorts/|live/)|youtu\.be/)([\w-]{11})(?:[?&].*)?$`)
	matches := re.FindStringSubmatch(videoURL)
	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", fmt.Errorf("URL de YouTube invalida: %s", videoURL)
}
