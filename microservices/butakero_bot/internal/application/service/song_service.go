package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
	"regexp"
	"time"
)

type songService struct {
	songRepo        ports.SongRepository
	messageProducer ports.MessageProducer
	messageConsumer ports.MessageConsumer
	logger          logging.Logger
}

func NewSongService(
	songRepo ports.SongRepository,
	messageProducer ports.MessageProducer,
	messageConsumer ports.MessageConsumer,
	logger logging.Logger,
) ports.SongService {
	return &songService{
		songRepo:        songRepo,
		messageProducer: messageProducer,
		messageConsumer: messageConsumer,
		logger:          logger,
	}
}

func (s *songService) extractURLOrTitle(input string) (string, bool) {
	urlRegex := regexp.MustCompile(`^(https?://)?(www\.)?(youtube\.com|youtu\.?be)/.+$`)
	return input, urlRegex.MatchString(input)
}

func (s *songService) GetOrDownloadSong(ctx context.Context, interactionID, userID, songInput, providerType string) (*entity.DiscordEntity, error) {
	input, isURL := s.extractURLOrTitle(songInput)
	s.logger.Info("Procesando solicitud de canción",
		zap.String("interactionID", interactionID),
		zap.String("userID", userID),
		zap.String("input", input),
		zap.Bool("isURL", isURL))

	// Primero intentamos buscar la canción en la base de datos
	var song *entity.SongEntity
	var err error

	if isURL {
		videoID := extractVideoID(input)
		if videoID != "" {
			song, err = s.songRepo.GetSongByVideoID(ctx, videoID)
			if err == nil && song != nil {
				s.logger.Info("Canción encontrada en la base de datos por videoID",
					zap.String("videoID", videoID))
				return songEntityToDiscordEntity(song), nil
			}
		}
	} else {
		songs, err := s.songRepo.SearchSongsByTitle(ctx, input)
		if err == nil && len(songs) > 0 {
			s.logger.Info("Canción encontrada en la base de datos por título",
				zap.String("title", songs[0].Metadata.Title))
			return songEntityToDiscordEntity(songs[0]), nil
		}
	}

	s.logger.Info("Canción no encontrada en DB, enviando solicitud a través de Kafka",
		zap.String("input", input))

	message := &entity.SongRequestMessage{
		InteractionID: interactionID,
		UserID:        userID,
		Song:          input,
		ProviderType:  providerType,
		Timestamp:     time.Now(),
	}

	err = s.messageProducer.PublishSongRequest(ctx, message)
	if err != nil {
		s.logger.Error("Error al publicar mensaje en Kafka",
			zap.String("input", input),
			zap.Error(err))
		return nil, fmt.Errorf("error al solicitar la descarga: %w", err)
	}

	s.logger.Info("Solicitud enviada a través de Kafka, esperando respuesta",
		zap.String("interactionID", interactionID),
		zap.String("song", input))

	msgChan := s.messageConsumer.GetMessagesChannel()

	for {
		select {
		case msg := <-msgChan:
			if msg.InteractionID != interactionID {
				s.logger.Warn("Mensaje recibido no corresponde a la operación actual",
					zap.String("esperado", interactionID),
					zap.String("recibido", msg.InteractionID))
				continue
			}

			if msg.Status == "success" {
				s.logger.Info("Descarga completada exitosamente",
					zap.String("interactionID", interactionID),
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
					zap.String("interactionID", interactionID),
					zap.String("error", msg.Message))
				return nil, fmt.Errorf("error en la descarga: %s", msg.Message)
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func songEntityToDiscordEntity(song *entity.SongEntity) *entity.DiscordEntity {
	return &entity.DiscordEntity{
		TitleTrack:   song.Metadata.Title,
		DurationMs:   song.Metadata.DurationMs,
		Platform:     song.Metadata.Platform,
		FilePath:     song.FileData.FilePath,
		ThumbnailURL: song.Metadata.ThumbnailURL,
		URL:          song.Metadata.URL,
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
