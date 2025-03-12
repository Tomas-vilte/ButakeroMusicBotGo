package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
)

type AudioStorageService struct {
	storage ports.Storage
	log     logger.Logger
}

func NewAudioStorageService(storage ports.Storage, logger logger.Logger) *AudioStorageService {
	return &AudioStorageService{
		storage: storage,
		log:     logger,
	}
}

func (as *AudioStorageService) StoreAudio(ctx context.Context, buffer *bytes.Buffer, songName string) (*model.FileData, error) {
	log := as.log.With(
		zap.String("component", "AudioStorageService"),
		zap.String("method", "StoreAudio"),
		zap.String("songName", songName),
	)
	keyName := fmt.Sprintf("%s%s", songName, ".dca")

	if err := as.storage.UploadFile(ctx, keyName, buffer); err != nil {
		log.Error("Error al subir el archivo", zap.Error(err))
		return nil, errors.ErrUploadFailed.Wrap(err)
	}

	fileData, err := as.storage.GetFileMetadata(ctx, keyName)
	if err != nil {
		log.Error("Error al obtener metadatos del archivo", zap.Error(err))
		return nil, errors.ErrUploadFailed.Wrap(err)
	}

	log.Info("Archivo de audio almacenado exitosamente", zap.String("key", keyName))
	return fileData, nil
}
