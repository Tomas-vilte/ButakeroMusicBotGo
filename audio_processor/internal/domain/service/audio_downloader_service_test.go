package service

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"strings"
	"testing"
)

func TestAudioDownloaderService_DownloadAndEncode_Success(t *testing.T) {
	ctx := context.Background()
	testURL := "https://test.com/audio.mp3"
	testAudioContent := "test audio content"
	testEncodeOptions := model.StdEncodeOptions

	mockDownloader := new(MockDownloader)
	mockEncoder := new(MockAudioEncoder)
	mockEncodeSession := new(MockEncodeSession)
	mockLogger := new(logger.MockLogger)

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	mockDownloader.On("DownloadAudio", ctx, testURL).Return(
		strings.NewReader(testAudioContent), nil)

	mockEncoder.On("Encode", ctx, mock.AnythingOfType("*strings.Reader"), testEncodeOptions).Return(
		mockEncodeSession, nil)

	frame1 := []byte("frame1")
	frame2 := []byte("frame2")

	mockEncodeSession.On("ReadFrame").Return(frame1, nil).Once()
	mockEncodeSession.On("ReadFrame").Return(frame2, nil).Once()
	mockEncodeSession.On("ReadFrame").Return([]byte{}, io.EOF).Once()
	mockEncodeSession.On("Cleanup").Return()

	service := NewAudioDownloaderService(mockDownloader, mockEncoder, mockLogger, testEncodeOptions)

	buffer, err := service.DownloadAndEncode(ctx, testURL)

	assert.NoError(t, err)
	assert.NotNil(t, buffer)
	assert.Equal(t, append(frame1, frame2...), buffer.Bytes())

	mockDownloader.AssertExpectations(t)
	mockEncoder.AssertExpectations(t)
	mockEncodeSession.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestAudioDownloaderService_DownloadAndEncode_DownloadError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	testURL := "https://test.com/invalid.mp3"
	testEncodeOptions := model.StdEncodeOptions
	expectedErr := errors.New("download failed")

	mockDownloader := new(MockDownloader)
	mockEncoder := new(MockAudioEncoder)
	mockLogger := new(logger.MockLogger)

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockDownloader.On("DownloadAudio", ctx, testURL).Return(nil, expectedErr)

	service := NewAudioDownloaderService(mockDownloader, mockEncoder, mockLogger, testEncodeOptions)

	buffer, err := service.DownloadAndEncode(ctx, testURL)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, buffer)

	mockDownloader.AssertExpectations(t)
	mockEncoder.AssertNotCalled(t, "Encode")
	mockLogger.AssertExpectations(t)
}

func TestAudioDownloaderService_DownloadAndEncode_EncodeError(t *testing.T) {
	ctx := context.Background()
	testURL := "https://test.com/audio.mp3"
	testAudioContent := "test audio content"
	testEncodeOptions := model.StdEncodeOptions
	expectedErr := errors.New("encode failed")

	mockDownloader := new(MockDownloader)
	mockEncoder := new(MockAudioEncoder)
	mockLogger := new(logger.MockLogger)

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockDownloader.On("DownloadAudio", ctx, testURL).Return(
		strings.NewReader(testAudioContent), nil)

	mockEncoder.On("Encode", ctx, mock.AnythingOfType("*strings.Reader"), testEncodeOptions).Return(
		nil, expectedErr)

	service := NewAudioDownloaderService(mockDownloader, mockEncoder, mockLogger, testEncodeOptions)

	buffer, err := service.DownloadAndEncode(ctx, testURL)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, buffer)

	mockDownloader.AssertExpectations(t)
	mockEncoder.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestAudioDownloaderService_DownloadAndEncode_ReadFrameError(t *testing.T) {
	ctx := context.Background()
	testURL := "https://test.com/audio.mp3"
	testAudioContent := "test audio content"
	testEncodeOptions := model.StdEncodeOptions
	expectedErr := errors.New("read frame failed")

	mockDownloader := new(MockDownloader)
	mockEncoder := new(MockAudioEncoder)
	mockEncodeSession := new(MockEncodeSession)
	mockLogger := new(logger.MockLogger)

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockDownloader.On("DownloadAudio", ctx, testURL).Return(
		strings.NewReader(testAudioContent), nil)

	mockEncoder.On("Encode", ctx, mock.AnythingOfType("*strings.Reader"), testEncodeOptions).Return(
		mockEncodeSession, nil)

	mockEncodeSession.On("ReadFrame").Return(nil, expectedErr)
	mockEncodeSession.On("Cleanup").Return()

	service := NewAudioDownloaderService(mockDownloader, mockEncoder, mockLogger, testEncodeOptions)

	buffer, err := service.DownloadAndEncode(ctx, testURL)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, buffer)

	mockDownloader.AssertExpectations(t)
	mockEncoder.AssertExpectations(t)
	mockEncodeSession.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestAudioDownloaderService_DownloadAndEncode_MaxSizeExceeded(t *testing.T) {
	// Arrange
	ctx := context.Background()
	testURL := "https://test.com/audio.mp3"
	testAudioContent := "test audio content"
	testEncodeOptions := model.StdEncodeOptions

	mockDownloader := new(MockDownloader)
	mockEncoder := new(MockAudioEncoder)
	mockEncodeSession := new(MockEncodeSession)
	mockLogger := new(logger.MockLogger)

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockDownloader.On("DownloadAudio", ctx, testURL).Return(
		strings.NewReader(testAudioContent), nil)

	mockEncoder.On("Encode", ctx, mock.AnythingOfType("*strings.Reader"), testEncodeOptions).Return(
		mockEncodeSession, nil)

	largeFrame := make([]byte, 101*1024*1024)

	mockEncodeSession.On("ReadFrame").Return(largeFrame, nil)
	mockEncodeSession.On("Cleanup").Return()

	service := NewAudioDownloaderService(mockDownloader, mockEncoder, mockLogger, testEncodeOptions)

	buffer, err := service.DownloadAndEncode(ctx, testURL)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tamaño máximo de audio")
	assert.Nil(t, buffer)

	mockDownloader.AssertExpectations(t)
	mockEncoder.AssertExpectations(t)
	mockEncodeSession.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
