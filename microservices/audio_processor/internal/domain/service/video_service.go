package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
)

type VideoService struct {
	providers map[string]ports.VideoProvider
	log       logger.Logger
}

func NewVideoService(providers map[string]ports.VideoProvider, logger logger.Logger) *VideoService {
	return &VideoService{
		providers: providers,
		log:       logger,
	}
}

func (s *VideoService) GetMediaDetails(ctx context.Context, input string, providerType string) (*model.MediaDetails, error) {
	s.log.With(
		zap.String("component", "VideoService"),
		zap.String("method", "GetMediaDetails"),
		zap.String("input", input),
		zap.String("providerType", providerType),
	)

	providerTypeLower := strings.ToLower(providerType)
	provider, ok := s.providers[providerTypeLower]
	if !ok {
		s.log.Warn("Tipo de proveedor no soportado", zap.String("provider_type", providerType))
		return nil, errors.ErrProviderNotFound.Wrap(fmt.Errorf("proveedor no soportado: %s", providerType))
	}

	videoID, err := provider.SearchVideoID(ctx, input)
	if err != nil {
		s.log.Error("Error al buscar ID de video", zap.Error(err), zap.String("provider_type", providerType))
		return nil, errors.ErrExternalAPIError.Wrap(err)
	}

	s.log.Debug("Obteniendo detalles del medio", zap.String("provider_type", providerType), zap.String("video_id", videoID))
	mediaDetails, err := provider.GetVideoDetails(ctx, videoID)
	if err != nil {
		s.log.Error("Error al obtener detalles del medio", zap.Error(err), zap.String("provider_type", providerType), zap.String("video_id", videoID))
		return nil, errors.ErrExternalAPIError.Wrap(err)
	}
	return mediaDetails, nil
}
