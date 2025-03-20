package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"strings"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
)

type videoService struct {
	providers map[string]ports.VideoProvider
	log       logger.Logger
}

func NewVideoService(providers map[string]ports.VideoProvider, logger logger.Logger) ports.VideoService {
	return &videoService{
		providers: providers,
		log:       logger,
	}
}

func (s *videoService) GetMediaDetails(ctx context.Context, input string, providerType string) (*model.MediaDetails, error) {
	log := s.log.With(
		zap.String("component", "VideoService"),
		zap.String("method", "GetMediaDetails"),
		zap.String("input", input),
		zap.String("providerType", providerType),
	)

	providerTypeLower := strings.ToLower(providerType)
	provider, ok := s.providers[providerTypeLower]
	if !ok {
		log.Warn("Tipo de proveedor no soportado", zap.String("provider_type", providerType))
		return nil, errors.ErrProviderNotFound.WithMessage(fmt.Sprintf("Proveedor no encontrado: %s", providerType))
	}

	videoID, err := provider.SearchVideoID(ctx, input)
	if err != nil {
		log.Error("Error al buscar ID de video", zap.Error(err), zap.String("provider_type", providerType))
		return nil, err
	}

	log.Debug("Obteniendo detalles de la cancion", zap.String("provider_type", providerType), zap.String("video_id", videoID))
	mediaDetails, err := provider.GetVideoDetails(ctx, videoID)
	if err != nil {
		log.Error("Error al obtener detalles del medio", zap.Error(err), zap.String("provider_type", providerType), zap.String("video_id", videoID))
		return nil, err
	}
	return mediaDetails, nil
}
