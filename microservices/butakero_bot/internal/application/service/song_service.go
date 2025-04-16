package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"regexp"
	"time"
)

type songService struct {
	mediaClient     ports.MediaClient
	messageProducer ports.MessageProducer
	messageConsumer ports.MessageConsumer
	logger          logging.Logger
}

func NewSongService(
	mediaClient ports.MediaClient,
	messageProducer ports.MessageProducer,
	messageConsumer ports.MessageConsumer,
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

func (s *songService) GetOrDownloadSong(ctx context.Context, userID, songInput, providerType string) (*entity.DiscordEntity, error) {
	requestID := uuid.New().String()

	input, isURL := s.extractURLOrTitle(songInput)
	s.logger.Info("Procesando solicitud de canción",
		zap.String("requestID", requestID),
		zap.String("userID", userID),
		zap.String("input", input),
		zap.Bool("isURL", isURL))

	var media *model.Media
	var err error

	if isURL {
		videoID := extractVideoID(input)
		if videoID != "" {
			media, err = s.mediaClient.GetMediaByID(ctx, videoID)
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

	s.logger.Info("Media no encontrada a través de la API, enviando solicitud a través de la queue",
		zap.String("input", input))

	message := &entity.SongRequestMessage{
		RequestID:    requestID,
		UserID:       userID,
		Song:         input,
		ProviderType: providerType,
		Timestamp:    time.Now(),
	}

	if err := s.messageProducer.PublishSongRequest(ctx, message); err != nil {
		s.logger.Error("Error al publicar mensaje en la queue",
			zap.String("input", input),
			zap.Error(err))
		return nil, fmt.Errorf("error al solicitar la descarga: %w", err)
	}
	s.logger.Info("Solicitud enviada a través de la queue, esperando respuesta",
		zap.String("requestID", requestID),
		zap.String("song", input))

	msgChan := s.messageConsumer.GetMessagesChannel()

	for {
		select {
		case msg := <-msgChan:
			if msg.RequestID == requestID {
				s.logger.Info("Mensaje recibido para solicitud actual",
					zap.String("requestID", requestID),
					zap.String("status", msg.Status))
			}

			if msg.Status == "success" {
				s.logger.Info("Descarga completada exitosamente",
					zap.String("requestID", requestID),
					zap.String("video_id", msg.VideoID))
				return &entity.DiscordEntity{
					TitleTrack:   msg.PlatformMetadata.Title,
					DurationMs:   msg.PlatformMetadata.DurationMs,
					ThumbnailURL: msg.PlatformMetadata.ThumbnailURL,
					Platform:     msg.PlatformMetadata.Platform,
					FilePath:     msg.FileData.FilePath,
					URL:          msg.PlatformMetadata.URL,
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
	// Patrones para diferentes formatos de URL de YouTube
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
