package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"strings"
)

type VideoService struct {
	youtubeProvider ports.VideoProvider
	spotifyProvider ports.VideoProvider
	log             logger.Logger
}

func NewVideoService(youtubeProvider ports.VideoProvider, spotifyProvider ports.VideoProvider, logger logger.Logger) *VideoService {
	return &VideoService{
		youtubeProvider: youtubeProvider,
		spotifyProvider: spotifyProvider,
		log:             logger,
	}
}

func (s *VideoService) GetMediaDetails(ctx context.Context, input string, providerType string) (*model.MediaDetails, error) {
	s.log.With(
		zap.String("component", "VideoService"),
		zap.String("method", "GetMediaDetails"),
		zap.String("input", input),
		zap.String("providerType", providerType),
	)

	var provider ports.VideoProvider
	var videoID string
	var err error

	providerTypeLower := strings.ToLower(providerType)

	switch providerTypeLower {
	case "youtube":
		provider = s.youtubeProvider
		videoID, err = provider.SearchVideoID(ctx, input)
		if err != nil {
			s.log.Error("Error al buscar ID de video en YouTube", zap.Error(err))
			return nil, errors.ErrExternalAPIError.Wrap(err)
		}
	case "spotify":
		provider = s.spotifyProvider
		videoID, err = provider.SearchVideoID(ctx, input)
		if err != nil {
			s.log.Error("Error al buscar ID de canción en Spotify", zap.Error(err))
			return nil, errors.ErrExternalAPIError.Wrap(err)
		}
	default:
		s.log.Warn("Tipo de proveedor no soportado", zap.String("provider_type", providerType))
		return nil, errors.ErrProviderNotFound.Wrap(fmt.Errorf("proveedor no soportado: %s", providerType))
	}

	s.log.Debug("Obteniendo detalles del medio", zap.String("provider_type", providerType), zap.String("video_id", videoID))
	mediaDetails, err := provider.GetVideoDetails(ctx, videoID)
	if err != nil {
		s.log.Error("Error al obtener detalles del medio", zap.Error(err), zap.String("provider_type", providerType), zap.String("video_id", videoID))
		return nil, errors.ErrExternalAPIError.Wrap(err)
	}
	return mediaDetails, nil
}
