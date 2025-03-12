package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
)

// MediaService es la implementación de ports.MediaService.
type MediaService struct {
	repo   ports.MediaRepository
	logger logger.Logger
}

// NewMediaService crea una nueva instancia de MediaService.
func NewMediaService(repo ports.MediaRepository, logger logger.Logger) *MediaService {
	return &MediaService{
		repo:   repo,
		logger: logger,
	}
}

// CreateMedia crea un nuevo registro de Media.
func (s *MediaService) CreateMedia(ctx context.Context, media *model.Media) error {
	log := s.logger.With(
		zap.String("component", "MediaService"),
		zap.String("method", "CreateMedia"),
		zap.String("media_id", media.ID),
	)

	if err := s.repo.SaveMedia(ctx, media); err != nil {
		log.Error("Error al crear el registro de media", zap.Error(err))
		return fmt.Errorf("error al crear el registro de media: %w", err)
	}

	log.Info("Registro de media creado exitosamente")
	return nil
}

// GetMediaByID obtiene un registro de Media por su ID y videoID.
func (s *MediaService) GetMediaByID(ctx context.Context, id, videoID string) (*model.Media, error) {
	log := s.logger.With(
		zap.String("component", "MediaService"),
		zap.String("method", "GetMediaByID"),
		zap.String("media_id", id),
		zap.String("video_id", videoID),
	)

	media, err := s.repo.GetMedia(ctx, id, videoID)
	if err != nil {
		log.Error("Error al obtener el registro de media", zap.Error(err))
		return nil, fmt.Errorf("error al obtener el registro de media: %w", err)
	}

	log.Info("Registro de media obtenido exitosamente")
	return media, nil
}

// UpdateMedia actualiza el estado de un registro de Media.
func (s *MediaService) UpdateMedia(ctx context.Context, id, videoID string, media *model.Media) error {
	log := s.logger.With(
		zap.String("component", "MediaService"),
		zap.String("method", "UpdateMediaStatus"),
		zap.String("media_id", id),
		zap.String("video_id", videoID),
	)

	if err := s.repo.UpdateMedia(ctx, id, videoID, media); err != nil {
		log.Error("Error al actualizar el estado del registro de media", zap.Error(err))
		return fmt.Errorf("error al actualizar el estado del registro de media: %w", err)
	}

	log.Info("Estado del registro de media actualizado exitosamente")
	return nil
}

// DeleteMedia elimina un registro de Media por su ID y videoID.
func (s *MediaService) DeleteMedia(ctx context.Context, id, videoID string) error {
	log := s.logger.With(
		zap.String("component", "MediaService"),
		zap.String("method", "DeleteMedia"),
		zap.String("media_id", id),
		zap.String("video_id", videoID),
	)

	if err := s.repo.DeleteMedia(ctx, id, videoID); err != nil {
		log.Error("Error al eliminar el registro de media", zap.Error(err))
		return fmt.Errorf("error al eliminar el registro de media: %w", err)
	}

	log.Info("Registro de media eliminado exitosamente")
	return nil
}
