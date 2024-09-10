package unit

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/storage"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"strings"
	"testing"
)

type mockPutObjectAPI func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)

func (m mockPutObjectAPI) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m(ctx, params, optFns...)
}

func TestS3Storage_UploadFile(t *testing.T) {
	t.Run("Successful upload", func(t *testing.T) {
		// arrange
		mockClient := mockPutObjectAPI(func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			// assert
			if params.Bucket == nil {
				t.Fatal("se espera que el bucket no sea nulo")
			}
			if got, want := *params.Bucket, "test-bucket"; got != want {
				t.Errorf("bucket = %q, se esperaba %q", got, want)
			}
			if params.Key == nil {
				t.Fatal("se espera que la clave no sea nula")
			}
			if got, want := *params.Key, "audio/test-file.txt"; got != want {
				t.Errorf("clave = %q, se esperaba %q", got, want)
			}
			return &s3.PutObjectOutput{}, nil
		})

		storageS3 := storage.S3Storage{
			Client:     mockClient,
			BucketName: "test-bucket",
		}

		// act
		err := storageS3.UploadFile(context.Background(), "test-file.txt", strings.NewReader("test content"))

		// assert
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
	})

	t.Run("Upload error", func(t *testing.T) {
		// arrange
		expectedErr := errors.New("s3 error")
		mockClient := mockPutObjectAPI(func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return nil, expectedErr
		})

		storageS3 := storage.S3Storage{
			Client:     mockClient,
			BucketName: "test-bucket",
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
	})

	t.Run("Nil Body", func(t *testing.T) {
		// arrange
		storageS3 := storage.S3Storage{
			Client: mockPutObjectAPI(func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				t.Fatal("No se debería llamar a PutObject")
				return nil, nil
			}),
			BucketName: "test-bucket",
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
	})
}

func TestNewS3Storage(t *testing.T) {
	t.Run("Successful creation", func(t *testing.T) {
		// act
		storageS3, err := storage.NewS3Storage("test-bucket", "us-east-1")

		// assert
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		if storageS3 == nil {
			t.Fatal("se esperaba un storage no nulo")
		}
		if got, want := storageS3.BucketName, "test-bucket"; got != want {
			t.Errorf("bucketName = %q, se esperaba %q", got, want)
		}
	})
}
