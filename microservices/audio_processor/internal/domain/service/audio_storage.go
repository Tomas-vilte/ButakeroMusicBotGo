package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
)

type AudioStorage struct {
	storage       ports.Storage
	metadataStore ports.MetadataRepository
	log           logger.Logger
}

func NewAudioStorage(s ports.Storage, m ports.MetadataRepository, l logger.Logger) *AudioStorage {
	return &AudioStorage{
		storage:       s,
		metadataStore: m,
		log:           l,
	}
}

func (as *AudioStorage) StoreAudio(ctx context.Context, buffer *bytes.Buffer, metadata *model.Metadata) (*model.FileData, error) {
	keyName := fmt.Sprintf("%s%s", metadata.Title, audioFileExtension)

	if err := as.storage.UploadFile(ctx, keyName, buffer); err != nil {
		return nil, errors.ErrUploadFailed.Wrap(err)
	}

	fileData, err := as.storage.GetFileMetadata(ctx, keyName)
	if err != nil {
		return nil, errors.ErrUploadFailed.Wrap(err)
	}

	if err := as.metadataStore.SaveMetadata(ctx, metadata); err != nil {
		return nil, errors.ErrMetadataSaveFailed.Wrap(err)
	}
	return fileData, nil
}
