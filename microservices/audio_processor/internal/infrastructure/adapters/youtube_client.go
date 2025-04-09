package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	errorsApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	youtubeBaseURL = "https://youtube.googleapis.com/youtube/v3"
	defaultTimeout = 10 * time.Second
)

type YouTubeClient struct {
	ApiKey     string
	BaseURL    string
	HttpClient *http.Client
	log        logger.Logger
}

func NewYouTubeClient(apiKey string, log logger.Logger) *YouTubeClient {
	return &YouTubeClient{
		ApiKey:  apiKey,
		BaseURL: youtubeBaseURL,
		HttpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		log: log,
	}
}

func (c *YouTubeClient) GetVideoDetails(ctx context.Context, videoID string) (*model.MediaDetails, error) {
	log := c.log.With(
		zap.String("component", "YouTubeClient"),
		zap.String("video_id", videoID),
		zap.String("method", "GetVideoDetails"),
	)
	log.Debug("Iniciando la obtención de detalles del video")

	if !isValidVideoID(videoID) {
		return nil, errorsApp.ErrCodeInvalidVideoID.WithMessage(fmt.Sprintf("ID de video inválido: %s", videoID))
	}

	endpoint := fmt.Sprintf("%s/videos?part=snippet,contentDetails&id=%s&key=%s", c.BaseURL, videoID, c.ApiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		log.Error("Error al crear la solicitud HTTP", zap.Error(err))
		return nil, errorsApp.ErrCodeGetVideoDetailsFailed.Wrap(err)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		log.Error("Error al ejecutar la solicitud HTTP", zap.Error(err))
		return nil, errorsApp.ErrYouTubeAPIError.WithMessage(fmt.Sprintf("Error al hacer la solicitud a la API de YouTube: %v", err))
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error("Error al cerrar el body de la respuesta", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Error("Error en la API de YouTube", zap.Int("status_code", resp.StatusCode))

		var youtubeError struct {
			Error struct {
				Message string `json:"message"`
				Errors  []struct {
					Message  string `json:"message"`
					Domain   string `json:"domain"`
					Reason   string `json:"reason"`
					Location string `json:"location"`
				} `json:"errors"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&youtubeError); err == nil {
			if len(youtubeError.Error.Errors) > 0 {
				errorDetails := make([]string, 0)
				for _, e := range youtubeError.Error.Errors {
					errorDetails = append(errorDetails, fmt.Sprintf("domain: %s, reason: %s, message: %s", e.Domain, e.Reason, e.Message))
				}
				return nil, errorsApp.ErrYouTubeAPIError.WithMessage(fmt.Sprintf("API de YouTube respondió con código %d: %s. Detalles: %v", resp.StatusCode, youtubeError.Error.Message, strings.Join(errorDetails, "; ")))
			}
			return nil, errorsApp.ErrYouTubeAPIError.WithMessage(fmt.Sprintf("API de YouTube respondió con código %d: %s", resp.StatusCode, youtubeError.Error.Message))
		}
		return nil, errorsApp.ErrYouTubeAPIError.WithMessage(fmt.Sprintf("API de YouTube respondió con código %d", resp.StatusCode))
	}

	var result struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Thumbnails struct {
					Default struct {
						URL string `json:"url"`
					} `json:"default"`
					MaxRes struct {
						URL string `json:"url"`
					} `json:"maxres"`
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

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error("Error al decodificar la respuesta de la API de YouTube", zap.Error(err))
		return nil, errorsApp.ErrCodeGetVideoDetailsFailed.WithMessage(fmt.Sprintf("Error al decodificar la respuesta de la API de YouTube: %v", err))
	}

	if len(result.Items) == 0 {
		log.Warn("No se encontró el video con el ID proporcionado")
		return nil, errorsApp.ErrCodeMediaNotFound.WithMessage(fmt.Sprintf("No se encontró el video con el ID %s", videoID))
	}

	item := result.Items[0]
	publishedAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
	if err != nil {
		log.Error("Error al parsear la fecha de publicación", zap.Error(err))
		return nil, errorsApp.ErrCodeGetVideoDetailsFailed.WithMessage(fmt.Sprintf("Error al parsear la fecha de publicación: %v", err))
	}

	durationMs, err := parseISODurationToMs(item.ContentDetails.Duration)
	if err != nil {
		log.Error("Error al convertir la duración", zap.Error(err))
		return nil, errorsApp.ErrCodeGetVideoDetailsFailed.WithMessage(fmt.Sprintf("Error al convertir la duración: %v", err))
	}

	thumbnailURL := item.Snippet.Thumbnails.MaxRes.URL
	if thumbnailURL == "" {
		thumbnailURL = item.Snippet.Thumbnails.Default.URL
	}

	videoDetails := &model.MediaDetails{
		Title:        item.Snippet.Title,
		ID:           item.ID,
		Description:  item.Snippet.Description,
		Creator:      item.Snippet.ChannelTitle,
		DurationMs:   durationMs,
		ThumbnailURL: thumbnailURL,
		PublishedAt:  publishedAt,
		URL:          fmt.Sprintf("https://youtube.com/watch?v=%s", videoID),
		Provider:     "YouTube",
	}
	log.Debug("Detalles del video obtenidos correctamente", zap.String("video_title", videoDetails.Title))
	return videoDetails, nil
}

func (c *YouTubeClient) SearchVideoID(ctx context.Context, input string) (string, error) {
	log := c.log.With(
		zap.String("component", "YouTubeClient"),
		zap.String("input", input),
		zap.String("method", "SearchVideoID"),
	)
	log.Info("Buscando ID del video")

	if strings.Contains(input, "youtube.com/watch") || strings.Contains(input, "youtu.be/") {
		log.Debug("La entrada es una URL, extrayendo el ID")
		videoID, err := ExtractVideoIDFromURL(input)
		if err != nil {
			log.Error("Error al extraer ID del video de la URL", zap.Error(err), zap.String("url", input))
			return "", errorsApp.ErrCodeSearchVideoIDFailed.WithMessage(fmt.Sprintf("Error al extraer ID del video de la URL: %v", err))
		}
		log.Debug("ID del video extraído de la URL", zap.String("video_id", videoID))
		return videoID, nil
	}

	encodedQuery := url.QueryEscape(input)
	endpoint := fmt.Sprintf("%s/search?part=id&q=%s&key=%s&type=video&maxResults=1", c.BaseURL, encodedQuery, c.ApiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		log.Error("Error al crear la solicitud HTTP", zap.Error(err))
		return "", errorsApp.ErrCodeSearchVideoIDFailed.Wrap(err)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		log.Error("Error al ejecutar la solicitud HTTP", zap.Error(err))
		return "", errorsApp.ErrYouTubeAPIError.WithMessage(fmt.Sprintf("Error al hacer la solicitud a la API de YouTube: %v", err))
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error("Error al cerrar el body de la respuesta", zap.Error(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Error("Error en la API de YouTube", zap.Int("status_code", resp.StatusCode))

		var youtubeError struct {
			Error struct {
				Message string `json:"message"`
				Errors  []struct {
					Message  string `json:"message"`
					Domain   string `json:"domain"`
					Reason   string `json:"reason"`
					Location string `json:"location"`
				} `json:"errors"`
			} `json:"error"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&youtubeError); err == nil {
			if len(youtubeError.Error.Errors) > 0 {
				errorDetails := make([]string, 0)
				for _, e := range youtubeError.Error.Errors {
					errorDetails = append(errorDetails, fmt.Sprintf("domain: %s, reason: %s, message: %s", e.Domain, e.Reason, e.Message))
				}
				return "", errorsApp.ErrYouTubeAPIError.WithMessage(fmt.Sprintf("API de YouTube respondió con código %d: %s. Detalles: %v", resp.StatusCode, youtubeError.Error.Message, strings.Join(errorDetails, "; ")))
			}
			return "", errorsApp.ErrYouTubeAPIError.WithMessage(fmt.Sprintf("API de YouTube respondió con código %d: %s", resp.StatusCode, youtubeError.Error.Message))
		}
		return "", errorsApp.ErrYouTubeAPIError.WithMessage(fmt.Sprintf("API de YouTube respondió con código %d", resp.StatusCode))
	}

	var result struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error("Error al decodificar la respuesta de la API de YouTube", zap.Error(err))
		return "", errorsApp.ErrCodeSearchVideoIDFailed.WithMessage(fmt.Sprintf("Error al decodificar la respuesta de la API de YouTube: %v", err))
	}

	if len(result.Items) == 0 {
		log.Warn("No se encontraron videos para la consulta", zap.String("input", input))
		return "", errorsApp.ErrCodeMediaNotFound.WithMessage(fmt.Sprintf("No se encontraron videos para la consulta: %s", input))
	}
	log.Debug("Video encontrado", zap.String("video_id", result.Items[0].ID.VideoID))
	return result.Items[0].ID.VideoID, nil
}

func ExtractVideoIDFromURL(videoURL string) (string, error) {
	re := regexp.MustCompile(`^(?:https?://)?(?:www\.)?(?:youtube\.com/(?:watch\?v=|embed/|v/|.+/(?:embed|v)/|shorts/|live/)|youtu\.be/)([\w-]{11})(?:[?&].*)?$`)
	matches := re.FindStringSubmatch(videoURL)
	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", errorsApp.ErrCodeInvalidVideoID.WithMessage(fmt.Sprintf("URL de YouTube inválida: %s", videoURL))
}

func isValidVideoID(videoID string) bool {
	return len(videoID) == 11
}

func parseISODurationToMs(isoDuration string) (int64, error) {
	durationStr := strings.TrimPrefix(isoDuration, "PT")
	durationStr = strings.ToLower(durationStr)
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, errorsApp.ErrCodeGetVideoDetailsFailed.WithMessage(fmt.Sprintf("Error al parsear la duración ISO 8601: %v", err))
	}
	return duration.Milliseconds(), nil
}
