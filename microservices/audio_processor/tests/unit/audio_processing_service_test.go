package unit

import (
	"bytes"
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"testing"
	"time"
)

func TestAudioProcessingService(t *testing.T) {
	setupTestAudioService := func() (*service.AudioProcessingService, *MockLogger, *MockStorage, *MockDownloader, *MockOperationRepository, *MockMetadataRepository, *MockMessagingQueue, *MockEncoder, *MockEncodeSession) {
		mockLogger := new(MockLogger)
		mockStorage := new(MockStorage)
		mockDownloader := new(MockDownloader)
		mockOperationRepo := new(MockOperationRepository)
		mockMetadataRepo := new(MockMetadataRepository)
		mockMessagingQueue := new(MockMessagingQueue)
		mockEncoder := new(MockEncoder)
		mockEncodeSession := new(MockEncodeSession)

		configService := &config.Config{
			Service: config.ServiceConfig{
				MaxAttempts: 1,
				Timeout:     2 * time.Second,
			},
		}

		serviceAudio := service.NewAudioProcessingService(
			mockLogger,
			mockStorage,
			mockDownloader,
			mockOperationRepo,
			mockMetadataRepo,
			mockMessagingQueue,
			mockEncoder,
			configService,
		)

		return serviceAudio, mockLogger, mockStorage, mockDownloader, mockOperationRepo, mockMetadataRepo, mockMessagingQueue, mockEncoder, mockEncodeSession
	}

	t.Run("StartOperation", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// Arrange
			serviceAudio, _, _, _, mockOperationRepo, _, _, _, _ := setupTestAudioService()

			ctx := context.Background()
			songID := "test-song-id"

			mockOperationRepo.On("SaveOperationsResult", mock.Anything, mock.AnythingOfType("*model.OperationResult")).Return(nil)

			// Act
			operationID, sk, err := serviceAudio.StartOperation(ctx, songID)

			// Assert
			assert.NoError(t, err)
			assert.NotEmpty(t, operationID)
			assert.Equal(t, songID, sk)
			mockOperationRepo.AssertExpectations(t)
		})
	})

	t.Run("FailureToSaveOperation", func(t *testing.T) {
		// Arrange
		serviceAudio, _, _, _, mockOperationRepo, _, _, _, _ := setupTestAudioService()

		ctx := context.Background()
		songID := "test-song-id"

		mockOperationRepo.On("SaveOperationsResult", mock.Anything, mock.AnythingOfType("*model.OperationResult")).Return(errors.New("database error"))

		// Act
		operationID, sk, err := serviceAudio.StartOperation(ctx, songID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, operationID)
		assert.Empty(t, sk)
		assert.Contains(t, err.Error(), "error al guardar resultado de operacion")
		mockOperationRepo.AssertExpectations(t)
	})

	t.Run("ProcessAudio", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// arrange
			serviceAudio, mockLogger, mockStorage, mockDownloader, mockOperationRepo, mockMetadataRepo, mockMessagingQueue, mockEncoder, mockEncodeSession := setupTestAudioService()

			ctx := context.Background()
			operationID := "test-operation-id"
			youtubeMetadata := &api.VideoDetails{
				VideoID:    "test-video-id",
				Title:      "Test Video",
				Duration:   "3:00",
				URLYouTube: "https://youtube.com/watch?v=test-video-id",
				Thumbnail:  "https://img.youtube.com/vi/test-video-id/0.jpg",
			}

			mockAudioContent := bytes.NewReader([]byte("fake audio data"))
			audioFrame := []byte("mocked frame")

			mockDownloader.On("DownloadAudio", mock.Anything, mock.AnythingOfType("string")).Return(mockAudioContent, nil)
			mockEncoder.On("Encode", mock.Anything, mockAudioContent, mock.Anything).Return(mockEncodeSession, nil)

			mockEncodeSession.On("ReadFrame").Return(audioFrame, nil).Once()
			mockEncodeSession.On("ReadFrame").Return([]byte(nil), io.EOF)
			mockEncodeSession.On("Cleanup").Return()

			mockStorage.On("UploadFile", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
			mockStorage.On("GetFileMetadata", mock.Anything, mock.Anything).Return(&model.FileData{
				FilePath: "Test Video.dca",
				FileSize: "1024MB",
			}, nil)
			mockMetadataRepo.On("SaveMetadata", mock.Anything, mock.AnythingOfType("*model.Metadata")).Return(nil)
			mockOperationRepo.On("UpdateOperationResult", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			mockMessagingQueue.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

			// Act
			err := serviceAudio.ProcessAudio(ctx, operationID, youtubeMetadata)

			// Assert
			assert.NoError(t, err)
			mockDownloader.AssertExpectations(t)
			mockEncoder.AssertExpectations(t)
			mockEncodeSession.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
			mockMetadataRepo.AssertExpectations(t)
			mockOperationRepo.AssertExpectations(t)
			mockMessagingQueue.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})

		t.Run("EncodingError", func(t *testing.T) {
			// arrange
			serviceAudio, mockLogger, _, mockDownloader, mockOperationRepo, _, mockMessagingQueue, mockEncoder, mockEncodeSession := setupTestAudioService()

			ctx := context.Background()
			operationID := "test-operation-id"
			youtubeMetadata := &api.VideoDetails{
				VideoID:    "test-video-id",
				Title:      "Test Video",
				Duration:   "3:00",
				URLYouTube: "https://youtube.com/watch?v=test-video-id",
				Thumbnail:  "https://img.youtube.com/vi/test-video-id/0.jpg",
			}

			mockAudioContent := io.NopCloser(bytes.NewReader([]byte("fake audio data")))
			mockDownloader.On("DownloadAudio", mock.Anything, mock.AnythingOfType("string")).Return(mockAudioContent, nil)
			mockEncoder.On("Encode", mock.Anything, mockAudioContent, mock.Anything).Return(mockEncodeSession, errors.New("encoding error"))
			mockOperationRepo.On("UpdateOperationResult", mock.Anything, operationID, mock.AnythingOfType("*model.OperationResult")).Return(nil)
			mockMessagingQueue.On("SendMessage", mock.Anything, mock.Anything).Return(nil)
			mockLogger.On("Error", mock.Anything, mock.Anything).Return()

			// act
			err := serviceAudio.ProcessAudio(ctx, operationID, youtubeMetadata)

			// assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "procesamiento fallido después de varios intentos")
		})

		t.Run("ReadFrameError", func(t *testing.T) {
			// arrange
			serviceAudio, mockLogger, _, mockDownloader, mockOperationRepo, _, mockMessagingQueue, mockEncoder, mockEncodeSession := setupTestAudioService()

			ctx := context.Background()
			operationID := "test-operation-id"
			youtubeMetadata := &api.VideoDetails{
				VideoID:    "test-video-id",
				Title:      "Test Video",
				Duration:   "3:00",
				URLYouTube: "https://youtube.com/watch?v=test-video-id",
				Thumbnail:  "https://img.youtube.com/vi/test-video-id/0.jpg",
			}

			mockAudioContent := io.NopCloser(bytes.NewReader([]byte("fake audio data")))
			mockDownloader.On("DownloadAudio", mock.Anything, mock.AnythingOfType("string")).Return(mockAudioContent, nil)
			mockEncoder.On("Encode", mock.Anything, mockAudioContent, mock.Anything).Return(mockEncodeSession, nil)
			mockEncodeSession.On("ReadFrame").Return([]byte("frame error"), errors.New("frame read error"))
			mockEncodeSession.On("Cleanup").Return()
			mockOperationRepo.On("UpdateOperationResult", mock.Anything, operationID, mock.AnythingOfType("*model.OperationResult")).Return(nil)
			mockMessagingQueue.On("SendMessage", mock.Anything, mock.Anything).Return(nil)
			mockLogger.On("Error", mock.Anything, mock.Anything).Return()

			// act
			err := serviceAudio.ProcessAudio(ctx, operationID, youtubeMetadata)

			// assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "procesamiento fallido después de varios intentos")
		})

		t.Run("GetFileMetadataError", func(t *testing.T) {
			// arrange
			serviceAudio, mockLogger, mockStorage, mockDownloader, mockOperationRepo, _, mockMessagingQueue, mockEncoder, mockEncodeSession := setupTestAudioService()

			ctx := context.Background()
			operationID := "test-operation-id"
			youtubeMetadata := &api.VideoDetails{
				VideoID:    "test-video-id",
				Title:      "Test Video",
				Duration:   "3:00",
				URLYouTube: "https://youtube.com/watch?v=test-video-id",
				Thumbnail:  "https://img.youtube.com/vi/test-video-id/0.jpg",
			}

			mockAudioContent := io.NopCloser(bytes.NewReader([]byte("fake audio data")))
			mockDownloader.On("DownloadAudio", mock.Anything, mock.AnythingOfType("string")).Return(mockAudioContent, nil)
			mockEncoder.On("Encode", mock.Anything, mockAudioContent, mock.Anything).Return(mockEncodeSession, nil)
			mockEncodeSession.On("ReadFrame").Return([]byte("fake frame data"), nil).Once()
			mockEncodeSession.On("ReadFrame").Return([]byte(nil), io.EOF)
			mockEncodeSession.On("Cleanup").Return()
			mockStorage.On("UploadFile", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
			mockStorage.On("GetFileMetadata", mock.Anything, mock.Anything).Return(&model.FileData{}, errors.New("error al obtener metadata"))
			mockOperationRepo.On("UpdateOperationResult", mock.Anything, operationID, mock.AnythingOfType("*model.OperationResult")).Return(nil)
			mockMessagingQueue.On("SendMessage", mock.Anything, mock.Anything).Return(nil)
			mockLogger.On("Error", mock.Anything, mock.Anything).Return()

			// act
			err := serviceAudio.ProcessAudio(ctx, operationID, youtubeMetadata)

			// assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "procesamiento fallido después de varios intentos")
		})

		t.Run("UpdateOperationResultError", func(t *testing.T) {
			// arrange
			serviceAudio, mockLogger, mockStorage, mockDownloader, mockOperationRepo, mockMetadataRepo, _, mockEncoder, mockEncodeSession := setupTestAudioService()

			ctx := context.Background()
			operationID := "test-operation-id"
			youtubeMetadata := &api.VideoDetails{
				VideoID:    "test-video-id",
				Title:      "Test Video",
				Duration:   "3:00",
				URLYouTube: "https://youtube.com/watch?v=test-video-id",
				Thumbnail:  "https://img.youtube.com/vi/test-video-id/0.jpg",
			}

			mockAudioContent := io.NopCloser(bytes.NewReader([]byte("fake audio data")))
			mockDownloader.On("DownloadAudio", mock.Anything, mock.AnythingOfType("string")).Return(mockAudioContent, nil)
			mockEncoder.On("Encode", mock.Anything, mockAudioContent, mock.Anything).Return(mockEncodeSession, nil)
			mockEncodeSession.On("ReadFrame").Return([]byte("mocked frame"), nil).Once()
			mockEncodeSession.On("ReadFrame").Return([]byte(nil), io.EOF)
			mockEncodeSession.On("Cleanup").Return()
			mockStorage.On("UploadFile", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
			mockStorage.On("GetFileMetadata", mock.Anything, mock.Anything).Return(&model.FileData{}, nil)
			mockMetadataRepo.On("SaveMetadata", mock.Anything, mock.AnythingOfType("*model.Metadata")).Return(nil)
			mockOperationRepo.On("UpdateOperationResult", mock.Anything, operationID, mock.AnythingOfType("*model.OperationResult")).Return(errors.New("error actualizar operacion"))
			mockLogger.On("Error", mock.Anything, mock.Anything).Return()

			// act
			err := serviceAudio.ProcessAudio(ctx, operationID, youtubeMetadata)

			// assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "error al guardar resultado de operación")
		})

		t.Run("SendMessageError", func(t *testing.T) {
			// arrange
			serviceAudio, mockLogger, mockStorage, mockDownloader, mockOperationRepo, mockMetadataRepo, mockMessagingQueue, mockEncoder, mockEncodeSession := setupTestAudioService()

			ctx := context.Background()
			operationID := "test-operation-id"
			youtubeMetadata := &api.VideoDetails{
				VideoID:    "test-video-id",
				Title:      "Test Video",
				Duration:   "3:00",
				URLYouTube: "https://youtube.com/watch?v=test-video-id",
				Thumbnail:  "https://img.youtube.com/vi/test-video-id/0.jpg",
			}

			mockAudioContent := io.NopCloser(bytes.NewReader([]byte("fake audio data")))
			mockDownloader.On("DownloadAudio", mock.Anything, mock.AnythingOfType("string")).Return(mockAudioContent, nil)
			mockEncoder.On("Encode", mock.Anything, mockAudioContent, mock.Anything).Return(mockEncodeSession, nil)
			mockEncodeSession.On("ReadFrame").Return([]byte("mocked frame"), nil).Once()
			mockEncodeSession.On("ReadFrame").Return([]byte(nil), io.EOF)
			mockEncodeSession.On("Cleanup").Return()
			mockStorage.On("UploadFile", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
			mockStorage.On("GetFileMetadata", mock.Anything, mock.Anything).Return(&model.FileData{}, nil)
			mockMetadataRepo.On("SaveMetadata", mock.Anything, mock.AnythingOfType("*model.Metadata")).Return(nil)
			mockOperationRepo.On("UpdateOperationResult", mock.Anything, operationID, mock.AnythingOfType("*model.OperationResult")).Return(nil)
			mockMessagingQueue.On("SendMessage", mock.Anything, mock.Anything).Return(errors.New("error al enviar el mensaje"))
			mockLogger.On("Error", mock.Anything, mock.Anything).Return()

			// act
			err := serviceAudio.ProcessAudio(ctx, operationID, youtubeMetadata)

			// assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "error al enviar el mensaje")
		})
	})

	t.Run("DownloadError", func(t *testing.T) {
		// Arrange
		serviceAudio, mockLogger, _, mockDownloader, mockOperationRepo, _, mockMessagingQueue, _, _ := setupTestAudioService()

		ctx := context.Background()
		operationID := "test-operation-id"
		youtubeMetadata := &api.VideoDetails{
			VideoID:    "test-video-id",
			Title:      "Test Video",
			Duration:   "3:00",
			URLYouTube: "https://youtube.com/watch?v=test-video-id",
			Thumbnail:  "https://img.youtube.com/vi/test-video-id/0.jpg",
		}
		mockAudioContent := io.NopCloser(bytes.NewBufferString("fake audio content"))

		mockDownloader.On("DownloadAudio", mock.Anything, mock.AnythingOfType("string")).Return(mockAudioContent, errors.New("download error"))
		mockOperationRepo.On("UpdateOperationResult", mock.Anything, operationID, mock.AnythingOfType("*model.OperationResult")).Return(nil)
		mockMessagingQueue.On("SendMessage", mock.Anything, mock.Anything).Return(nil)
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()

		// Act
		err := serviceAudio.ProcessAudio(ctx, operationID, youtubeMetadata)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "procesamiento fallido después de varios intentos")
		mockDownloader.AssertExpectations(t)
		mockOperationRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("UploadError", func(t *testing.T) {
		// Arrange
		serviceAudio, mockLogger, mockStorage, mockDownloader, mockOperationRepo, mockMetadataRepo, mockMessagingQueue, mockEncoder, mockEncodeSession := setupTestAudioService()

		ctx := context.Background()
		operationID := "test-operation-id"
		youtubeMetadata := &api.VideoDetails{
			VideoID:    "test-video-id",
			Title:      "Test Video",
			Duration:   "3:00",
			URLYouTube: "https://youtube.com/watch?v=test-video-id",
			Thumbnail:  "https://img.youtube.com/vi/test-video-id/0.jpg",
		}

		mockAudioContent := io.NopCloser(bytes.NewBufferString("fake audio content"))
		audioFrame := []byte("mocked frame")

		mockDownloader.On("DownloadAudio", mock.Anything, mock.AnythingOfType("string")).Return(mockAudioContent, nil)
		mockEncoder.On("Encode", mock.Anything, mockAudioContent, mock.Anything).Return(mockEncodeSession, nil)
		mockEncodeSession.On("ReadFrame").Return(audioFrame, nil).Once()
		mockEncodeSession.On("ReadFrame").Return([]byte(nil), io.EOF)
		mockEncodeSession.On("Cleanup").Return()
		mockStorage.On("UploadFile", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(errors.New("error en subir el archivo"))
		mockMessagingQueue.On("SendMessage", mock.Anything, mock.Anything).Return(nil)
		mockOperationRepo.On("UpdateOperationResult", mock.Anything, operationID, mock.AnythingOfType("*model.OperationResult")).Return(nil)
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()

		// Act
		err := serviceAudio.ProcessAudio(ctx, operationID, youtubeMetadata)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "procesamiento fallido después de varios intentos")
		mockDownloader.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
		mockMetadataRepo.AssertExpectations(t)
		mockOperationRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("SaveMetadataError", func(t *testing.T) {
		// Arrange
		serviceAudio, mockLogger, mockStorage, mockDownloader, mockOperationRepo, mockMetadataRepo, mockMessagingQueue, mockEncoder, mockEncodeSession := setupTestAudioService()

		ctx := context.Background()
		operationID := "test-operation-id"
		youtubeMetadata := &api.VideoDetails{
			VideoID:    "test-video-id",
			Title:      "Test Video",
			Duration:   "3:00",
			URLYouTube: "https://youtube.com/watch?v=test-video-id",
			Thumbnail:  "https://img.youtube.com/vi/test-video-id/0.jpg",
		}

		mockAudioContent := io.NopCloser(bytes.NewBufferString("fake audio content"))
		audioFrame := []byte("mocked frame")

		mockDownloader.On("DownloadAudio", mock.Anything, mock.AnythingOfType("string")).Return(mockAudioContent, nil)
		mockEncoder.On("Encode", mock.Anything, mockAudioContent, mock.Anything).Return(mockEncodeSession, nil)
		mockEncodeSession.On("ReadFrame").Return(audioFrame, nil).Once()
		mockEncodeSession.On("ReadFrame").Return([]byte(nil), io.EOF)
		mockEncodeSession.On("Cleanup").Return()
		mockStorage.On("UploadFile", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
		mockStorage.On("GetFileMetadata", mock.Anything, mock.Anything).Return(&model.FileData{}, nil)
		mockMetadataRepo.On("SaveMetadata", mock.Anything, mock.AnythingOfType("*model.Metadata")).Return(errors.New("metadata save error"))
		mockOperationRepo.On("UpdateOperationResult", mock.Anything, operationID, mock.AnythingOfType("*model.OperationResult")).Return(nil)
		mockMessagingQueue.On("SendMessage", mock.Anything, mock.Anything).Return(nil)
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()

		// Act
		err := serviceAudio.ProcessAudio(ctx, operationID, youtubeMetadata)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "procesamiento fallido después de varios intentos")
		mockDownloader.AssertExpectations(t)
		mockEncoder.AssertExpectations(t)
		mockEncodeSession.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
		mockMetadataRepo.AssertExpectations(t)
		mockOperationRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
