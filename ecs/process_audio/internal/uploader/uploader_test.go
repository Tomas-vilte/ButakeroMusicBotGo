package uploader

import (
	"bytes"
	"context"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/ecs/process_audio/internal/config"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"testing"
)

func TestUploadToS3_Success(t *testing.T) {
	mockUploader := new(MockS3Uploader)
	mockS3Client := new(MockS3Client)
	mockLogger := new(MockLogger)

	cfg := config.Config{
		BucketName: "test-bucket",
		Region:     "us-east-1",
		AccessKey:  "test-access-key",
		SecretKey:  "test-secret-key",
	}

	s3Uploader := NewS3Uploader(mockS3Client, mockUploader, mockLogger, cfg)

	ctx := context.Background()
	audioData := bytes.NewReader([]byte("test audio data"))
	key := "test-key"

	mockUploader.On("UploadWithContext", mock.Anything, mock.Anything, mock.Anything).Return(&s3manager.UploadOutput{}, nil)
	mockLogger.On("Info", mock.Anything, mock.Anything)

	err := s3Uploader.UploadToS3(ctx, audioData, key)

	assert.NoError(t, err)
	mockUploader.AssertExpectations(t)
	mockLogger.AssertExpectations(t)

}

func TestUploadToS3_Failure(t *testing.T) {
	mockUploader := new(MockS3Uploader)
	mockS3Client := new(MockS3Client)
	mockLogger := new(MockLogger)

	cfg := config.Config{
		BucketName: "test-bucket",
		Region:     "us-west-2",
		AccessKey:  "test-access-key",
		SecretKey:  "test-secret-key",
	}

	s3Uploader := NewS3Uploader(mockS3Client, mockUploader, mockLogger, cfg)

	ctx := context.Background()
	audioData := bytes.NewReader([]byte("test audio data"))
	key := "test-key"

	mockUploader.On("UploadWithContext", mock.Anything, mock.Anything, mock.Anything).Return(&s3manager.UploadOutput{}, errors.New("upload error"))
	mockLogger.On("Info", mock.Anything, mock.Anything)
	mockLogger.On("Error", mock.Anything, mock.Anything)

	err := s3Uploader.UploadToS3(ctx, audioData, key)

	assert.Error(t, err)
	assert.Equal(t, "error al subir los datos DCA a S3", err.Error())
	mockUploader.AssertExpectations(t)
	mockLogger.AssertExpectations(t)

}

func TestDownloadFromS3_Success(t *testing.T) {
	mockUploader := new(MockS3Uploader)
	mockS3Client := new(MockS3Client)
	mockLogger := new(MockLogger)

	cfg := config.Config{
		BucketName: "test-bucket",
		Region:     "us-east-1",
		AccessKey:  "test-access-key",
		SecretKey:  "test-secret-key",
	}

	s3Uploader := NewS3Uploader(mockS3Client, mockUploader, mockLogger, cfg)

	ctx := context.Background()
	key := "test-key"

	mockS3Client.On("GetObjectWithContext", mock.Anything, mock.Anything, mock.Anything).Return(&s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewReader([]byte("test audio data"))),
	}, nil)

	result, err := s3Uploader.DownloadFromS3(ctx, key)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockS3Client.AssertExpectations(t)
}

func TestDownloadFromS3_Failure(t *testing.T) {
	mockUploader := new(MockS3Uploader)
	mockS3Client := new(MockS3Client)
	mockLogger := new(MockLogger)

	cfg := config.Config{
		BucketName: "test-bucket",
		Region:     "us-east-1",
		AccessKey:  "test-access-key",
		SecretKey:  "test-secret-key",
	}

	s3Uploader := NewS3Uploader(mockS3Client, mockUploader, mockLogger, cfg)

	ctx := context.Background()
	key := "test-key"

	mockS3Client.On("GetObjectWithContext", mock.Anything, mock.Anything, mock.Anything).Return(&s3.GetObjectOutput{}, errors.New("download error"))
	mockLogger.On("Error", mock.Anything, mock.Anything)

	result, err := s3Uploader.DownloadFromS3(ctx, key)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error al descargar el archivo de S3")
	mockS3Client.AssertExpectations(t)
	mockLogger.AssertExpectations(t)

}
