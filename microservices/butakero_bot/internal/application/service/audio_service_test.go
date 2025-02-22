package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
)

func TestAudioService_RequestDownload_Success(t *testing.T) {
	// Arrange
	mockDownloader := new(MockSongDownloader)
	loggerMock := new(logging.MockLogger)

	expectedResponse := &entity.DownloadResponse{
		OperationID: "op123",
		SongID:      "song456",
	}

	mockDownloader.On("DownloadSong", mock.Anything, "test-song").Return(expectedResponse, nil)
	loggerMock.On("Info", "Solicitud de descarga exitosa", []zap.Field{
		zap.String("songName", "test-song"),
		zap.String("operationId", "op123"),
	})

	service := NewAudioService(mockDownloader, loggerMock)

	// Act
	resp, err := service.RequestDownload(context.Background(), "test-song")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedResponse, resp)
	mockDownloader.AssertExpectations(t)
	loggerMock.AssertExpectations(t)
}

func TestAudioService_RequestDownload_DownloadError(t *testing.T) {
	// Arrange
	mockDownloader := new(MockSongDownloader)
	loggerMock := new(logging.MockLogger)

	expectedError := errors.New("download failed")
	mockDownloader.On("DownloadSong", mock.Anything, "bad-song").Return(&entity.DownloadResponse{}, expectedError)
	loggerMock.On("Error", "Error al solicitar la descarga", []zap.Field{
		zap.String("songName", "bad-song"),
		zap.Error(expectedError),
	})

	service := NewAudioService(mockDownloader, loggerMock)

	// Act
	_, err := service.RequestDownload(context.Background(), "bad-song")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error al solicitar la descarga")
	assert.Contains(t, err.Error(), expectedError.Error())
	mockDownloader.AssertExpectations(t)
	loggerMock.AssertExpectations(t)
}

func TestAudioService_RequestDownload_LoggingVerification(t *testing.T) {
	// Arrange
	mockDownloader := new(MockSongDownloader)
	loggerMock := new(logging.MockLogger)

	t.Run("error logging parameters", func(t *testing.T) {
		customError := fmt.Errorf("custom error")
		mockDownloader.On("DownloadSong", mock.Anything, "specific-song").Return(&entity.DownloadResponse{}, customError)
		loggerMock.On("Error", "Error al solicitar la descarga", []zap.Field{
			zap.String("songName", "specific-song"),
			zap.Error(customError),
		})

		service := NewAudioService(mockDownloader, loggerMock)

		// Act
		_, err := service.RequestDownload(context.Background(), "specific-song")

		// Assert
		assert.Error(t, err)
		loggerMock.AssertCalled(t, "Error", "Error al solicitar la descarga", []zap.Field{
			zap.String("songName", "specific-song"),
			zap.Error(customError),
		})
	})
}
