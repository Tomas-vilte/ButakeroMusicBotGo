//go:build !integration

package service

import (
	"bytes"
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAudioStorage_StoreAudio(t *testing.T) {
	t.Run("should store audio successfully", func(t *testing.T) {
		// arrange
		mockStorage := new(MockStorage)
		mockMetadataRepo := new(MockMetadataRepository)
		mockLogger := new(logger.MockLogger)

		audioStorage := NewAudioStorage(mockStorage, mockMetadataRepo, mockLogger)

		ctx := context.Background()
		buffer := bytes.NewBuffer([]byte("audio data"))
		metadata := &model.Metadata{
			Title: "test_audio",
		}

		mockStorage.On("UploadFile", ctx, "test_audio.dca", buffer).Return(nil)
		mockStorage.On("GetFileMetadata", ctx, "test_audio.dca").Return(&model.FileData{}, nil)
		mockMetadataRepo.On("SaveMetadata", ctx, metadata).Return(nil)

		// Act
		fileData, err := audioStorage.StoreAudio(ctx, buffer, metadata)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, fileData)
		mockStorage.AssertExpectations(t)
		mockMetadataRepo.AssertExpectations(t)
	})

	t.Run("should return error when upload fails", func(t *testing.T) {
		// arrange

		mockStorage := new(MockStorage)
		mockMetadataRepo := new(MockMetadataRepository)
		mockLogger := new(logger.MockLogger)

		audioStorage := NewAudioStorage(mockStorage, mockMetadataRepo, mockLogger)

		ctx := context.Background()
		buffer := bytes.NewBuffer([]byte("audio data"))
		metadata := &model.Metadata{
			Title: "test_audio",
		}

		mockStorage.On("UploadFile", ctx, "test_audio.dca", buffer).Return(errors.New("upload failed"))

		fileData, err := audioStorage.StoreAudio(ctx, buffer, metadata)

		// assert
		assert.Error(t, err)
		assert.Nil(t, fileData)
		mockStorage.AssertExpectations(t)

	})

	t.Run("should return error when getting file metadata fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(MockStorage)
		mockMetadataRepo := new(MockMetadataRepository)
		mockLogger := new(logger.MockLogger)

		audioStorage := NewAudioStorage(mockStorage, mockMetadataRepo, mockLogger)

		ctx := context.Background()
		buffer := bytes.NewBuffer([]byte("audio data"))
		metadata := &model.Metadata{
			Title: "test_audio",
		}

		mockStorage.On("UploadFile", ctx, "test_audio.dca", buffer).Return(nil)
		mockStorage.On("GetFileMetadata", ctx, "test_audio.dca").Return(&model.FileData{}, errors.New("metadata fetch failed"))

		// Act
		fileData, err := audioStorage.StoreAudio(ctx, buffer, metadata)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, fileData)
		mockStorage.AssertExpectations(t)
	})

	t.Run("should return error when saving metadata fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(MockStorage)
		mockMetadataRepo := new(MockMetadataRepository)
		mockLogger := new(logger.MockLogger)

		audioStorage := NewAudioStorage(mockStorage, mockMetadataRepo, mockLogger)

		ctx := context.Background()
		buffer := bytes.NewBuffer([]byte("audio data"))
		metadata := &model.Metadata{
			Title: "test_audio",
		}

		mockStorage.On("UploadFile", ctx, "test_audio.dca", buffer).Return(nil)
		mockStorage.On("GetFileMetadata", ctx, "test_audio.dca").Return(&model.FileData{}, nil)
		mockMetadataRepo.On("SaveMetadata", ctx, metadata).Return(errors.New("metadata save failed"))

		// Act
		fileData, err := audioStorage.StoreAudio(ctx, buffer, metadata)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, fileData)
		mockStorage.AssertExpectations(t)
		mockMetadataRepo.AssertExpectations(t)
	})
}
