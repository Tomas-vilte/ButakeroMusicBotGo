package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
)

type externalServiceAudio struct {
	logger     logging.Logger
	downloader ports.SongDownloader
}

func NewExternalAudioService(downloader ports.SongDownloader, logger logging.Logger) ports.ExternalSongService {
	return &externalServiceAudio{
		logger:     logger,
		downloader: downloader,
	}
}

func (e *externalServiceAudio) RequestDownload(ctx context.Context, songName, providerType string) (*entity.DownloadResponse, error) {
	response, err := e.downloader.DownloadSong(ctx, songName, providerType)
	if err != nil {
		e.logger.Error("Error al solicitar la descarga",
			zap.String("songName", songName),
			zap.Error(err))
		return nil, fmt.Errorf("%s", err)
	}

	e.logger.Info("Solicitud de descarga exitosa",
		zap.String("songName", songName),
		zap.String("operationId", response.OperationID))

	return response, nil
}
