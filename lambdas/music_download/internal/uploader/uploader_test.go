package uploader

import (
	"bytes"
	"context"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/config"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestS3Uploader_UploadDCA(t *testing.T) {
	mockUploader := new(MockS3Uploader)
	mockLogger := new(MockLogger)
	mockConfig := &config.Config{
		BucketName: "test-bucket",
	}

	s3Uploader := &S3Uploader{
		S3Uploader: mockUploader,
		Logger:     mockLogger,
		Config:     *mockConfig,
	}

	audioData := bytes.NewReader([]byte("audio data"))
	key := "test-key.dca"

	mockUploader.On("UploadWithContext", mock.Anything, mock.Anything).Return(&s3manager.UploadOutput{}, nil)
	mockLogger.On("Info", "Iniciando carga de datos DCA a S3", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Info", "Datos DCA subidos exitosamente a S3", mock.AnythingOfType("[]zapcore.Field")).Return()

	err := s3Uploader.UploadToS3(context.Background(), audioData, key)

	// Assert
	assert.NoError(t, err)
	mockUploader.AssertExpectations(t)

}

func TestS3Uploader_UploadDCA_Error(t *testing.T) {
	// Arrange
	mockUploader := new(MockS3Uploader)
	mockLogger := new(MockLogger)
	mockConfig := &config.Config{
		BucketName: "test-bucket",
	}

	s3Uploader := &S3Uploader{
		S3Uploader: mockUploader,
		Logger:     mockLogger,
		Config:     *mockConfig,
	}

	audioData := bytes.NewReader([]byte("audio data"))
	key := "test-key.dca"

	expectedError := errors.New("error al subir los datos DCA a S3")
	mockUploader.On("UploadWithContext", mock.Anything, mock.Anything).Return(&s3manager.UploadOutput{}, expectedError)
	mockLogger.On("Info", "Iniciando carga de datos DCA a S3", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Error", "Error al subir los datos DCA a S3", mock.AnythingOfType("[]zapcore.Field")).Return()

	// Act
	err := s3Uploader.UploadToS3(context.Background(), audioData, key)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockUploader.AssertExpectations(t)
}
