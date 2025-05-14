package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model/queue"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"regexp"
	"time"
)

type songService struct {
	mediaClient     ports.MediaClient
	messageProducer ports.SongDownloadRequestPublisher
	messageConsumer ports.SongDownloadEventSubscriber
	logger          logging.Logger
}

func NewSongService(
	mediaClient ports.MediaClient,
	messageProducer ports.SongDownloadRequestPublisher,
	messageConsumer ports.SongDownloadEventSubscriber,
	logger logging.Logger,
) ports.SongService {
	return &songService{
		mediaClient:     mediaClient,
		messageProducer: messageProducer,
		messageConsumer: messageConsumer,
		logger:          logger,
	}
}

func (s *songService) extractURLOrTitle(input string) (string, bool) {
	urlRegex := regexp.MustCompile(`^(https?://)?(www\.)?(youtube\.com|youtu\.?be)/.+$`)
	return input, urlRegex.MatchString(input)
}

func (s *songService) GetSongFromAPI(ctx context.Context, input string) (*entity.DiscordEntity, error) {
	_, isURL := s.extractURLOrTitle(input)

	if isURL {
		videoID := extractVideoID(input)
		if videoID != "" {
			media, err := s.mediaClient.GetMediaByID(ctx, videoID)
			if err == nil && media != nil {
				s.logger.Info("Media encontrada a través de la API por videoID",
					zap.String("videoID", videoID))
				return mediaToDiscordEntity(media), nil
			}
		}
	} else {
		mediaList, err := s.mediaClient.SearchMediaByTitle(ctx, input)
		if err == nil && len(mediaList) > 0 {
			s.logger.Info("Media encontrada a través de la API por título",
				zap.String("title", mediaList[0].Metadata.Title))
			return mediaToDiscordEntity(mediaList[0]), nil
		}
	}

	return nil, fmt.Errorf("canción no encontrada en la API")
}

func (s *songService) DownloadSongViaQueue(ctx context.Context, userID, input, providerType string) (*entity.DiscordEntity, error) {
	requestID := uuid.New().String()

	s.logger.Info("Enviando solicitud a través de la queue",
		zap.String("input", input),
		zap.String("requestID", requestID),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("userID", userID))

	message := &queue.DownloadRequestMessage{
		RequestID:    requestID,
		UserID:       userID,
		Song:         input,
		ProviderType: providerType,
		Timestamp:    time.Now(),
	}

	if err := s.messageProducer.PublishDownloadRequest(ctx, message); err != nil {
		s.logger.Error("Error al publicar mensaje en la queue",
			zap.String("input", input),
			zap.Error(err))
		return nil, fmt.Errorf("error al solicitar la descarga: %w", err)
	}

	return s.waitForDownloadResponse(ctx, requestID)
}

func (s *songService) waitForDownloadResponse(ctx context.Context, requestID string) (*entity.DiscordEntity, error) {
	msgChan := s.messageConsumer.DownloadEventsChannel()

	for {
		select {
		case msg := <-msgChan:
			if msg.RequestID != requestID {
				continue
			}

			s.logger.Info("Mensaje recibido para solicitud actual",
				zap.String("requestID", requestID),
				zap.String("status", msg.Status))

			if msg.Status == "success" {
				s.logger.Info("Descarga completada exitosamente",
					zap.String("requestID", requestID),
					zap.String("video_id", msg.VideoID))
				return &entity.DiscordEntity{
					ID:           uuid.New().String(),
					TitleTrack:   msg.PlatformMetadata.Title,
					DurationMs:   msg.PlatformMetadata.DurationMs,
					ThumbnailURL: msg.PlatformMetadata.ThumbnailURL,
					Platform:     msg.PlatformMetadata.Platform,
					FilePath:     msg.FileData.FilePath,
					URL:          msg.PlatformMetadata.URL,
					AddedAt:      time.Now(),
				}, nil
			} else {
				s.logger.Error("Error en la descarga",
					zap.String("requestID", requestID),
					zap.String("error", msg.Message))
				return nil, fmt.Errorf("error en la descarga: %s", msg.Message)
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (s *songService) GetOrDownloadSong(ctx context.Context, userID, songInput, providerType string) (*entity.DiscordEntity, error) {
	song, err := s.GetSongFromAPI(ctx, songInput)
	if err == nil {
		return song, nil
	}

	s.logger.Info("Media no encontrada a través de la API, intentando descarga",
		zap.String("input", songInput),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("userID", userID))

	return s.DownloadSongViaQueue(ctx, userID, songInput, providerType)
}

func mediaToDiscordEntity(media *model.Media) *entity.DiscordEntity {
	return &entity.DiscordEntity{
		TitleTrack:   media.Metadata.Title,
		DurationMs:   media.Metadata.DurationMs,
		Platform:     media.Metadata.Platform,
		FilePath:     media.FileData.FilePath,
		ThumbnailURL: media.Metadata.ThumbnailURL,
		URL:          media.Metadata.URL,
	}
}

func extractVideoID(url string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`youtube\.com/watch\?v=([^&]+)`),
		regexp.MustCompile(`youtu\.be/([^?]+)`),
		regexp.MustCompile(`youtube\.com/embed/([^?]+)`),
	}

	for _, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(url); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}
