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

type SongService struct {
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
) *SongService {
	return &SongService{
		songRepo:        songRepo,
		externalService: externalService,
		messageConsumer: messageConsumer,
		logger:          logger,
	}
}

func (s *SongService) extractURLOrTitle(input string) (string, bool) {
	urlRegex := regexp.MustCompile(`^(https?://)?(www\.)?(youtube\.com|youtu\.?be)/.+$`)
	return input, urlRegex.MatchString(input)
}

func (s *SongService) GetOrDownloadSong(ctx context.Context, song, providerType string) (*entity.Song, error) {
	input, isURL := s.extractURLOrTitle(song)
	s.logger.Info("Procesando solicitud de canción",
		zap.String("input", input),
		zap.Bool("isURL", isURL))

	var songs []*entity.Song
	var err error

	if isURL {
		videoID := extractVideoID(input)
		song, err := s.songRepo.GetSongByVideoID(ctx, videoID)
		if err == nil && song != nil {
			return song, nil
		}
	} else {
		songs, err = s.songRepo.SearchSongsByTitle(ctx, input)
		if err == nil && len(songs) > 0 {
			return songs[0], nil
		}
	}

	s.logger.Info("Canción no encontrada en DB, iniciando descarga",
		zap.String("input", input))

	downloadResp, err := s.externalService.RequestDownload(ctx, input, providerType)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	msgChan := s.messageConsumer.GetMessagesChannel()
	go func() {
		if err := s.messageConsumer.ConsumeMessages(ctx, 0); err != nil {
			s.logger.Error("Error al consumir mensajes",
				zap.String("input", input),
				zap.Error(err))
		}
	}()

	for {
		select {
		case msg := <-msgChan:
			if msg.Status.ID != downloadResp.OperationID {
				s.logger.Warn("Mensaje recibido no corresponde a la operación actual",
					zap.String("esperado", downloadResp.OperationID),
					zap.String("recibido", msg.Status.ID))
				continue
			}

			if msg.Status.Status == "success" {
				s.logger.Info("Descarga completada exitosamente",
					zap.String("input", input))
				return &entity.Song{
					ID: msg.Status.ID,
				}, nil
			}
			return nil, fmt.Errorf("error en la descarga: %s", msg.Status.Message)
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
