//go:build !integration

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
		VideoID:  "video123",
		Status:   "success",
		Provider: "youtube",
		Success:  true,
	}

	mockDownloader.On("DownloadSong", mock.Anything, "video123", "youtube").Return(expectedResponse, nil)
	loggerMock.On("Info", mock.Anything, mock.Anything).Return()

	service := NewExternalAudioService(mockDownloader, loggerMock)

	// Act
	resp, err := service.RequestDownload(context.Background(), "video123", "youtube")

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
	mockDownloader.On("DownloadSong", mock.Anything, "bad-song", "youtube").Return(&entity.DownloadResponse{}, expectedError)
	loggerMock.On("Error", mock.Anything, mock.Anything).Return()

	service := NewExternalAudioService(mockDownloader, loggerMock)

	// Act
	_, err := service.RequestDownload(context.Background(), "bad-song", "youtube")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "download failed")
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
		mockDownloader.On("DownloadSong", mock.Anything, "specific-song", "youtube").Return(&entity.DownloadResponse{}, customError)
		loggerMock.On("Error", "Error al solicitar la descarga", []zap.Field{
			zap.String("songName", "specific-song"),
			zap.Error(customError),
		})

		service := NewExternalAudioService(mockDownloader, loggerMock)

		// Act
		_, err := service.RequestDownload(context.Background(), "specific-song", "youtube")

		// Assert
		assert.Error(t, err)
		loggerMock.AssertCalled(t, "Error", "Error al solicitar la descarga", []zap.Field{
			zap.String("songName", "specific-song"),
			zap.Error(customError),
		})
	})
}
