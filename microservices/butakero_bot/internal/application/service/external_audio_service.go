package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/errors_app"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
	"regexp"
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
		var appErr *errors_app.AppError
		if errors.As(err, &appErr) && appErr.Code == errors_app.ErrCodeAPIDuplicateRecord {
			e.logger.Info("El video ya está registrado, obteniendo el ID del video del mensaje de error",
				zap.String("songName", songName),
				zap.String("errorMessage", appErr.Message))

			videoID := extractVideoIDFromErrorMessage(appErr.Message)
			if videoID == "" {
				e.logger.Error("No se pudo extraer el ID del video del mensaje de error",
					zap.String("errorMessage", appErr.Message))
				return nil, fmt.Errorf("no se pudo extraer el ID del video del mensaje de error")
			}

			return &entity.DownloadResponse{
				Provider: providerType,
				Status:   "duplicate_record",
				Success:  false,
				VideoID:  videoID,
			}, nil
		}

		e.logger.Error("Error al solicitar la descarga",
			zap.String("songName", songName),
			zap.Error(err))
		return nil, fmt.Errorf("%s", err)
	}

	e.logger.Info("Solicitud de descarga exitosa",
		zap.String("songName", songName),
		zap.String("videoID", response.VideoID))

	return response, nil
}

// extractVideoIDFromErrorMessage extrae el ID del video del mensaje de error.
func extractVideoIDFromErrorMessage(errorMessage string) string {
	re := regexp.MustCompile(`'([^']+)'`)
	matches := re.FindStringSubmatch(errorMessage)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
