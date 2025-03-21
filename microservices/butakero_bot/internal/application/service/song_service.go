package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
	"regexp"
)

type songService struct {
	songRepo        ports.SongRepository
	externalService ports.ExternalSongService
	messageConsumer ports.MessageConsumer
	logger          logging.Logger
}

func NewSongService(
	songRepo ports.SongRepository,
	externalService ports.ExternalSongService,
	messageConsumer ports.MessageConsumer,
	logger logging.Logger,
) ports.SongService {
	return &songService{
		songRepo:        songRepo,
		externalService: externalService,
		messageConsumer: messageConsumer,
		logger:          logger,
	}
}

func (s *songService) extractURLOrTitle(input string) (string, bool) {
	urlRegex := regexp.MustCompile(`^(https?://)?(www\.)?(youtube\.com|youtu\.?be)/.+$`)
	return input, urlRegex.MatchString(input)
}

func (s *songService) GetOrDownloadSong(ctx context.Context, songInput, providerType string) (*entity.DiscordEntity, error) {
	input, isURL := s.extractURLOrTitle(songInput)
	s.logger.Info("Procesando solicitud de canción",
		zap.String("input", input),
		zap.Bool("isURL", isURL))

	var songs []*entity.SongEntity
	var err error

	if isURL {
		videoID := extractVideoID(input)
		song, err := s.songRepo.GetSongByVideoID(ctx, videoID)
		if err == nil && song != nil {
			return &entity.DiscordEntity{
				TitleTrack:   song.Metadata.Title,
				DurationMs:   song.Metadata.DurationMs,
				Platform:     song.Metadata.Platform,
				FilePath:     song.FileData.FilePath,
				ThumbnailURL: song.Metadata.ThumbnailURL,
			}, nil
		}
	} else {
		songs, err = s.songRepo.SearchSongsByTitle(ctx, input)
		if err == nil && len(songs) > 0 {
			return &entity.DiscordEntity{
				TitleTrack:   songs[0].Metadata.Title,
				DurationMs:   songs[0].Metadata.DurationMs,
				Platform:     songs[0].Metadata.Platform,
				FilePath:     songs[0].FileData.FilePath,
				ThumbnailURL: songs[0].Metadata.ThumbnailURL,
			}, nil
		}
	}

	s.logger.Info("Canción no encontrada en DB, iniciando descarga",
		zap.String("input", input))

	response, err := s.externalService.RequestDownload(ctx, input, providerType)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	if response.Status == "duplicate_record" {
		s.logger.Info("El video ya está registrado, consultando la base de datos",
			zap.String("videoID", response.VideoID))

		song, err := s.songRepo.GetSongByVideoID(ctx, response.VideoID)
		if err != nil {
			return nil, fmt.Errorf("error al obtener la canción de la base de datos: %s", err)
		}

		if song != nil {
			return &entity.DiscordEntity{
				TitleTrack:   song.Metadata.Title,
				DurationMs:   song.Metadata.DurationMs,
				Platform:     song.Metadata.Platform,
				FilePath:     song.FileData.FilePath,
				ThumbnailURL: song.Metadata.ThumbnailURL,
			}, nil
		}
	}

	if !response.Success {
		return nil, fmt.Errorf("la solicitud de descarga falló: %s", response.Status)
	}

	videoID := response.VideoID
	s.logger.Info("Solicitud de descarga enviada",
		zap.String("video_id", videoID),
		zap.String("provider", response.Provider),
		zap.String("status", response.Status))

	msgChan := s.messageConsumer.GetMessagesChannel()

	for {
		select {
		case msg := <-msgChan:
			if msg.VideoID != videoID {
				s.logger.Warn("Mensaje recibido no corresponde a la operación actual",
					zap.String("esperado", videoID),
					zap.String("recibido", msg.VideoID))
				continue
			}

			if msg.Status == "success" {
				s.logger.Info("Descarga completada exitosamente",
					zap.String("video_id", videoID))
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
					zap.String("video_id", videoID),
					zap.String("error", msg.Message))
				return nil, fmt.Errorf("error en la descarga: %s", msg.Message)
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
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
