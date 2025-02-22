package s3_storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
	"time"
)

var (
	testBucket  = os.Getenv("AWS_BUCKET_NAME")
	testPrefix  = "audio/"
	testTimeout = 30 * time.Second
	testConfig  = &config.Config{
		AWS: config.AWSConfig{
			Region: "us-east-1",
		},
		Storage: config.StorageConfig{
			S3Config: config.S3Config{
				BucketName: testBucket,
			},
		},
	}
)

func TestS3StorageIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Saltando prueba de integración en modo corto")
	}

	logger, err := logging.NewZapLogger()
	require.NoError(t, err)

	storage, err := NewS3Storage(testConfig, logger)
	require.NoError(t, err, "Error creando s3Storage")

	s3Client := storage.client

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	t.Run("Obtener audio existente", func(t *testing.T) {
		testKey := testPrefix + "testfile_" + randomString(8) + ".mp3"
		testContent := []byte("fake audio content " + randomString(20))

		_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(testBucket),
			Key:    aws.String(testKey),
			Body:   bytesReader(testContent),
		})
		require.NoError(t, err, "Error subiendo archivo de prueba")

		defer func() {
			_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(testBucket),
				Key:    aws.String(testKey),
			})
			require.NoError(t, err, "Error limpiando archivo de prueba")
		}()

		reader, err := storage.GetAudio(ctx, testKey)
		require.NoError(t, err, "Error obteniendo audio")
		defer reader.Close()

		content, err := io.ReadAll(reader)
		require.NoError(t, err, "Error leyendo contenido")

		require.Equal(t, testContent, content, "Contenido no coincide")
	})

	t.Run("Error en archivo inexistente", func(t *testing.T) {
		invalidKey := testPrefix + "non-existent-file_" + randomString(12) + ".mp3"

		_, err := storage.GetAudio(ctx, invalidKey)
		require.Error(t, err, "Debería generar error")
		require.Contains(t, err.Error(), "error al obtener de S3", "Mensaje de error incorrecto")
	})

	t.Run("Contexto cancelado", func(t *testing.T) {
		ctxCanceled, cancel := context.WithCancel(ctx)
		cancel()

		_, err := storage.GetAudio(ctxCanceled, "anyfile.mp3")
		require.ErrorIs(t, err, context.Canceled, "Debería detectar contexto cancelado")
	})
}

func bytesReader(data []byte) io.Reader {
	return bytes.NewReader(data)
}

func randomString(length int) string {
	b := make([]byte, length/2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
