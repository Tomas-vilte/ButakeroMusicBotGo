//go:build !integration

package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	errors2 "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestAudioStorageService_StoreAudio(t *testing.T) {
	// Arrange
	mockStorage := new(MockStorage)
	mockLogger := new(logger.MockLogger)

	service := NewAudioStorageService(mockStorage, mockLogger)

	songName := "test-song"
	keyName := fmt.Sprintf("%s%s", songName, ".dca")
	buffer := bytes.NewBuffer([]byte("test audio data"))
	expectedFileData := &model.FileData{
		FilePath: keyName,
		FileSize: "1234",
		FileType: "audio/dca",
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockStorage.On("UploadFile", mock.Anything, keyName, buffer).Return(nil)
	mockStorage.On("GetFileMetadata", mock.Anything, keyName).Return(expectedFileData, nil)

	// Act
	fileData, err := service.StoreAudio(context.Background(), buffer, songName)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedFileData, fileData)
	mockStorage.AssertExpectations(t)
}

func TestAudioStorageService_StoreAudio_UploadError(t *testing.T) {
	// Arrange
	mockStorage := new(MockStorage)
	mockLogger := new(logger.MockLogger)

	service := NewAudioStorageService(mockStorage, mockLogger)

	songName := "test-song"
	keyName := fmt.Sprintf("%s%s", songName, ".dca")
	buffer := bytes.NewBuffer([]byte("test audio data"))
	expectedError := errors.New("upload failed")

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockStorage.On("UploadFile", mock.Anything, keyName, buffer).Return(expectedError)

	// Act
	fileData, err := service.StoreAudio(context.Background(), buffer, songName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, fileData)

	var appErr *errors2.AppError
	ok := errors.As(err, &appErr)
	assert.True(t, ok, "El error debería ser de tipo *errors.AppError")
	assert.Equal(t, "upload_failed", appErr.Code)
	assert.Contains(t, err.Error(), expectedError.Error())

	mockStorage.AssertExpectations(t)
}

func TestAudioStorageService_StoreAudio_MetadataError(t *testing.T) {
	// Arrange
	mockStorage := new(MockStorage)
	mockLogger := new(logger.MockLogger)

	service := NewAudioStorageService(mockStorage, mockLogger)

	songName := "test-song"
	keyName := fmt.Sprintf("%s%s", songName, ".dca")
	buffer := bytes.NewBuffer([]byte("test audio data"))
	expectedError := errors.New("metadata failed")

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockStorage.On("UploadFile", mock.Anything, keyName, buffer).Return(nil)
	mockStorage.On("GetFileMetadata", mock.Anything, keyName).Return(&model.FileData{}, expectedError)

	// Act
	fileData, err := service.StoreAudio(context.Background(), buffer, songName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, fileData)

	var appErr *errors2.AppError
	ok := errors.As(err, &appErr)
	assert.True(t, ok, "El error debería ser de tipo *errors.AppError")
	assert.Equal(t, "upload_failed", appErr.Code)
	assert.Contains(t, "Error al obtener metadatos del archivo: metadata failed", appErr.Message)

	mockStorage.AssertExpectations(t)
}
