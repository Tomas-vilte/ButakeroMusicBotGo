package s3_audio

import (
	"bytes"
	"context"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/internal/config"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"testing"
)

func TestS3Uploader_UploadDCA(t *testing.T) {
	mockUploader := new(MockS3Uploader)
	mockLogger := new(logging.MockLogger)
	mockConfig := &config.Config{
		BucketName: "test-bucket",
	}

	s3Uploader := &S3Uploader{
		S3Uploader: mockUploader,
		Logger:     mockLogger,
		Config:     mockConfig,
	}

	audioData := bytes.NewReader([]byte("audio data"))
	key := "test-key.dca"

	mockUploader.On("UploadWithContext", mock.Anything, mock.Anything).Return(&s3manager.UploadOutput{}, nil)
	mockLogger.On("Info", "Iniciando carga de datos DCA a S3", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Info", "Datos DCA subidos exitosamente a S3", mock.AnythingOfType("[]zapcore.Field")).Return()

	err := s3Uploader.UploadDCA(context.Background(), audioData, key)

	// Assert
	assert.NoError(t, err)
	mockUploader.AssertExpectations(t)

}

func TestS3Uploader_FileExists(t *testing.T) {
	mockClient := new(MockS3Client)
	mockLogger := new(logging.MockLogger)
	mockConfig := &config.Config{
		BucketName: "test-bucket",
	}

	s3Uploader := &S3Uploader{
		S3Client: mockClient,
		Logger:   mockLogger,
		Config:   mockConfig,
	}

	key := "test-key.dca"

	mockClient.On("HeadObjectWithContext", mock.Anything, mock.Anything, mock.Anything).Return(&s3.HeadObjectOutput{}, nil)

	// Act
	exists, err := s3Uploader.FileExists(context.Background(), key)

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	mockClient.AssertExpectations(t)

}

func TestS3Uploader_DownloadDCA(t *testing.T) {
	mockDownloader := new(MockS3Downloader)
	mockLogger := new(logging.MockLogger)
	mockConfig := &config.Config{
		BucketName: "test-bucket",
	}

	s3Uploader := &S3Uploader{
		S3Downloader: mockDownloader,
		Logger:       mockLogger,
		Config:       mockConfig,
	}

	key := "test-key.dca"
	expectedData := []byte("downloaded data")
	mockDownloader.On("DownloadWithContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(len(expectedData)), nil).Run(func(args mock.Arguments) {
		writer := args.Get(1).(io.WriterAt)
		_, err := writer.WriteAt(expectedData, 0)
		if err != nil {
			return
		}
	})

	// Act
	result, err := s3Uploader.DownloadDCA(context.Background(), key)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	data, _ := io.ReadAll(result)
	assert.Equal(t, expectedData, data)
	mockDownloader.AssertExpectations(t)
}

func TestS3Uploader_UploadDCA_Error(t *testing.T) {
	// Arrange
	mockUploader := new(MockS3Uploader)
	mockLogger := new(logging.MockLogger)
	mockConfig := &config.Config{
		BucketName: "test-bucket",
	}

	s3Uploader := &S3Uploader{
		S3Uploader: mockUploader,
		Logger:     mockLogger,
		Config:     mockConfig,
	}

	audioData := bytes.NewReader([]byte("audio data"))
	key := "test-key.dca"

	expectedError := errors.New("error al subir los datos DCA a S3")
	mockUploader.On("UploadWithContext", mock.Anything, mock.Anything).Return(&s3manager.UploadOutput{}, expectedError)
	mockLogger.On("Info", "Iniciando carga de datos DCA a S3", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Error", "Error al subir los datos DCA a S3", mock.AnythingOfType("[]zapcore.Field")).Return()

	// Act
	err := s3Uploader.UploadDCA(context.Background(), audioData, key)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockUploader.AssertExpectations(t)
}

func TestS3Uploader_FileExists_Error(t *testing.T) {
	// Arrange
	mockClient := new(MockS3Client)
	mockLogger := new(logging.MockLogger)
	mockConfig := &config.Config{
		BucketName: "test-bucket",
	}

	s3Uploader := &S3Uploader{
		S3Client: mockClient,
		Logger:   mockLogger,
		Config:   mockConfig,
	}

	key := "test-key.dca"
	expectedError := errors.New("file exists error")
	mockClient.On("HeadObjectWithContext", mock.Anything, mock.Anything, mock.Anything).Return(&s3.HeadObjectOutput{}, expectedError)

	// Act
	exists, err := s3Uploader.FileExists(context.Background(), key)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.False(t, exists)
	mockClient.AssertExpectations(t)

}

func TestS3Uploader_DownloadDCA_Error(t *testing.T) {
	mockDownloader := new(MockS3Downloader)
	mockLogger := new(logging.MockLogger)
	mockConfig := &config.Config{
		BucketName: "test-bucket",
	}

	s3Uploader := &S3Uploader{
		S3Downloader: mockDownloader,
		Logger:       mockLogger,
		Config:       mockConfig,
	}

	key := "test-key.dca"
	expectedError := errors.New("download error")
	mockDownloader.On("DownloadWithContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(0), expectedError)

	// Act
	result, err := s3Uploader.DownloadDCA(context.Background(), key)
	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
	mockDownloader.AssertExpectations(t)
}
