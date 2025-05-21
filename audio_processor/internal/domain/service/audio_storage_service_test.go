//go:build !integration

package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
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

	songName := "testsong"
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

	songName := "testsong"
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

	assert.Equal(t, err.Error(), "upload failed")
	assert.Contains(t, err.Error(), expectedError.Error())

	mockStorage.AssertExpectations(t)
}

func TestAudioStorageService_StoreAudio_MetadataError(t *testing.T) {
	// Arrange
	mockStorage := new(MockStorage)
	mockLogger := new(logger.MockLogger)

	service := NewAudioStorageService(mockStorage, mockLogger)

	songName := "testsong"
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

	assert.Equal(t, err.Error(), "metadata failed")
	assert.Contains(t, "Error al obtener metadatos del archivo: metadata failed", err.Error())

	mockStorage.AssertExpectations(t)
}
