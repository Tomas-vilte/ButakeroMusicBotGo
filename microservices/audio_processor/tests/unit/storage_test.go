package unit

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/storage/cloud"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"testing"
)

func TestS3Storage_UploadFile(t *testing.T) {
	t.Run("Successful upload", func(t *testing.T) {
		// arrange
		mockClient := new(MockStorageS3API)
		mockClient.On("PutObject", mock.Anything, mock.AnythingOfType("*s3.PutObjectInput"), mock.Anything).
			Return(&s3.PutObjectOutput{}, nil)

		storageS3 := cloud.S3Storage{
			Client: mockClient,
			Config: config.Config{
				Storage: config.StorageConfig{
					S3Config: &config.S3Config{
						BucketName: "test-bucket",
					},
				},
			},
		}

		// act
		err := storageS3.UploadFile(context.Background(), "test-file.txt", strings.NewReader("test content"))

		// assert
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		mockClient.AssertExpectations(t)
	})

	t.Run("Upload error", func(t *testing.T) {
		// arrange
		expectedErr := errors.New("s3 error")
		mockClient := new(MockStorageS3API)
		mockClient.On("PutObject", mock.Anything, mock.AnythingOfType("*s3.PutObjectInput"), mock.Anything).
			Return((*s3.PutObjectOutput)(nil), expectedErr)

		storageS3 := cloud.S3Storage{
			Client: mockClient,
			Config: config.Config{
				Storage: config.StorageConfig{
					S3Config: &config.S3Config{
						BucketName: "test-bucket",
					},
				},
			},
		}

		// act
		err := storageS3.UploadFile(context.Background(), "test-file.txt", strings.NewReader("test content"))

		// assert
		if err == nil {
			t.Fatal("expected an error, but got none")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("got error %v, want %v", err, expectedErr)
		}
		mockClient.AssertExpectations(t)
	})

	t.Run("Nil Body", func(t *testing.T) {
		// arrange
		mockClient := new(MockStorageS3API)
		// No configuramos expectativas para PutObject porque no debería ser llamado

		storageS3 := cloud.S3Storage{
			Client: mockClient,
			Config: config.Config{
				Storage: config.StorageConfig{
					S3Config: &config.S3Config{
						BucketName: "test-bucket",
					},
				},
			},
		}

		// act
		err := storageS3.UploadFile(context.Background(), "test-file.txt", nil)

		// assert
		if err == nil {
			t.Fatal("se esperaba un error, pero no se obtuvo ninguno")
		}
		if got, want := err.Error(), "el cuerpo no puede ser nulo"; got != want {
			t.Errorf("error = %q, se esperaba %q", got, want)
		}
		mockClient.AssertNotCalled(t, "PutObject")
	})
}

func TestS3Storage_GetFileMetadata(t *testing.T) {
	t.Run("Successful metadata retrieval", func(t *testing.T) {
		// Arrange
		mockClient := new(MockStorageS3API)
		bucketName := "test-bucket"
		key := "test-file.dca"
		contentType := "application/octet-stream"
		contentLength := int64(1024 * 1024) // 1 MB

		mockClient.On("HeadObject", mock.Anything, &s3.HeadObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("audio/" + key),
		}, mock.Anything).Return(&s3.HeadObjectOutput{
			ContentType:   aws.String(contentType),
			ContentLength: aws.Int64(contentLength),
		}, nil)

		s3Storage := cloud.S3Storage{
			Client: mockClient,
			Config: config.Config{
				Storage: config.StorageConfig{
					S3Config: &config.S3Config{
						BucketName: "test-bucket",
					},
				},
			},
		}

		// Act
		fileData, err := s3Storage.GetFileMetadata(context.Background(), key)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, fileData)
		assert.Equal(t, "audio/"+key, fileData.FilePath)
		assert.Equal(t, contentType, fileData.FileType)
		assert.Equal(t, "1.00MB", fileData.FileSize)
		assert.Equal(t, fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, key), fileData.PublicURL)

		mockClient.AssertExpectations(t)
	})

	t.Run("Error retrieving metadata", func(t *testing.T) {
		// Arrange
		mockClient := new(MockStorageS3API)
		bucketName := "test-bucket"
		key := "non-existent-file.mp3"

		mockClient.On("HeadObject", mock.Anything, &s3.HeadObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("audio/" + key),
		}, mock.Anything).Return((*s3.HeadObjectOutput)(nil), errors.New("s3 error"))

		s3Storage := cloud.S3Storage{
			Client: mockClient,
			Config: config.Config{
				Storage: config.StorageConfig{
					S3Config: &config.S3Config{
						BucketName: "test-bucket",
					},
				},
			},
		}

		// Act
		fileData, err := s3Storage.GetFileMetadata(context.Background(), key)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, fileData)
		assert.Contains(t, err.Error(), "error obteniendo metadata del archivo de S3")

		mockClient.AssertExpectations(t)
	})
}

func TestNewS3Storage(t *testing.T) {
	t.Run("Successful creation", func(t *testing.T) {
		// act
		cfg := config.Config{
			Storage: config.StorageConfig{
				S3Config: &config.S3Config{
					BucketName: "test-bucket",
				},
			},
			AWS: &config.AWSConfig{
				Region: "us-east-1",
			},
		}
		storageS3, err := cloud.NewS3Storage(cfg)

		// assert
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		if storageS3 == nil {
			t.Fatal("se esperaba un storage no nulo")
		}
		if got, want := storageS3.Config.Storage.S3Config.BucketName, "test-bucket"; got != want {
			t.Errorf("bucketName = %q, se esperaba %q", got, want)
		}
	})
}
