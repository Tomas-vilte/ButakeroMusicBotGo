package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"strings"
	"time"
	"unicode"
)

type mediaService struct {
	repo   ports.MediaRepository
	logger logger.Logger
}

func NewMediaService(repo ports.MediaRepository, logger logger.Logger) ports.MediaService {
	return &mediaService{
		repo:   repo,
		logger: logger,
	}
}

func (s *mediaService) CreateMedia(ctx context.Context, media *model.Media) error {
	log := s.logger.With(
		zap.String("component", "MediaService"),
		zap.String("method", "CreateMedia"),
	)

	if err := s.validateMedia(media); err != nil {
		log.Error("Error de validación", zap.Error(err))
		return err
	}

	now := time.Now()
	media.CreatedAt = now
	media.UpdatedAt = now
	media.TitleLower = normalizedTitle(media.Metadata.Title)

	if err := s.repo.SaveMedia(ctx, media); err != nil {
		log.Error("Error al crear el registro de media", zap.Error(err))
		return err
	}
	log.Info("Registro de media creado exitosamente")
	return nil
}

func (s *mediaService) GetMediaByID(ctx context.Context, videoID string) (*model.Media, error) {
	log := s.logger.With(
		zap.String("component", "MediaService"),
		zap.String("method", "GetMediaByID"),
		zap.String("video_id", videoID),
	)

	media, err := s.repo.GetMedia(ctx, videoID)
	if err != nil {
		log.Error("Error al obtener el registro de media", zap.Error(err))
		return nil, err
	}
	log.Info("Registro de media obtenido exitosamente")
	return media, nil
}

func (s *mediaService) UpdateMedia(ctx context.Context, videoID string, media *model.Media) error {
	log := s.logger.With(
		zap.String("component", "MediaService"),
		zap.String("method", "UpdateMedia"),
		zap.String("video_id", videoID),
	)

	if err := s.validateMedia(media); err != nil {
		log.Error("Error de validación", zap.Error(err))
		return err
	}

	media.UpdatedAt = time.Now()
	media.TitleLower = normalizedTitle(media.Metadata.Title)
	media.VideoID = videoID

	if err := s.repo.UpdateMedia(ctx, videoID, media); err != nil {
		log.Error("Error al actualizar el registro de media", zap.Error(err))
		return err
	}
	log.Info("Registro de media actualizado exitosamente")
	return nil
}

func (s *mediaService) DeleteMedia(ctx context.Context, videoID string) error {
	log := s.logger.With(
		zap.String("component", "MediaService"),
		zap.String("method", "DeleteMedia"),
		zap.String("video_id", videoID),
	)

	if err := s.repo.DeleteMedia(ctx, videoID); err != nil {
		log.Error("Error al eliminar el registro de media", zap.Error(err))
		return err
	}
	log.Info("Registro de media eliminado exitosamente")
	return nil
}

func normalizedTitle(title string) string {
	normalized := strings.ToLower(title)

	normalized = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			return r
		}
		return -1
	}, normalized)

	return normalized
}

func (s *mediaService) validateMedia(media *model.Media) error {
	if media == nil {
		return fmt.Errorf("media no puede ser nil")
	}
	if media.Metadata.Title == "" {
		return fmt.Errorf("el título es requerido")
	}
	if media.VideoID == "" {
		return fmt.Errorf("el ID del video es requerido")
	}
	return nil
}
