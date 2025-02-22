package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
)

type audioService struct {
	logger     logging.Logger
	downloader ports.SongDownloader
}

func NewAudioService(downloader ports.SongDownloader, logger logging.Logger) ports.ExternalSongService {
	return &audioService{
		logger:     logger,
		downloader: downloader,
	}
}

func (s *audioService) RequestDownload(ctx context.Context, songName string) (*entity.DownloadResponse, error) {
	response, err := s.downloader.DownloadSong(ctx, songName)
	if err != nil {
		s.logger.Error("Error al solicitar la descarga",
			zap.String("songName", songName),
			zap.Error(err))
		return nil, fmt.Errorf("error al solicitar la descarga: %w", err)
	}

	s.logger.Info("Solicitud de descarga exitosa",
		zap.String("songName", songName),
		zap.String("operationId", response.OperationID))

	return response, nil
}
