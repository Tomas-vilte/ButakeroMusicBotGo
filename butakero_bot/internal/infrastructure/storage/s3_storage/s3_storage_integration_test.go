//go:build integration

package s3_storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"testing"
	"time"
)

const (
	testBucket  = "test-bucket"
	testPrefix  = "audio/"
	awsEndpoint = "http://localhost:4566"
	region      = "us-east-1"
)

type localstackContainer struct {
	testcontainers.Container
	optsFunc func(service, region string, options ...interface{}) (aws.Endpoint, error)
}

func setupLocalstack(ctx context.Context) (*localstackContainer, error) {
	port := "4566:4566/tcp"
	req := testcontainers.ContainerRequest{
		Image:        "localstack/localstack:latest",
		ExposedPorts: []string{port},
		WaitingFor: wait.ForAll(
			wait.ForLog("Ready"),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	optsFunc := func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           awsEndpoint,
			SigningRegion: region,
		}, nil
	}

	return &localstackContainer{
		Container: container,
		optsFunc:  optsFunc,
	}, nil
}

func setupS3Client(ctx context.Context, optsFunc func(service, region string, options ...interface{}) (aws.Endpoint, error)) (*s3.Client, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(optsFunc)

	awsCfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(region),
		awsConfig.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.UsePathStyle = true
	})

	return client, nil
}

func createBucket(ctx context.Context, client *s3.Client, bucket string) error {
	_, err := client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	return err
}

func TestS3StorageIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Saltando prueba de integración en modo corto")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	localstack, err := setupLocalstack(ctx)
	require.NoError(t, err, "Error iniciando contenedor de Localstack")
	defer func() {
		if err := localstack.Terminate(ctx); err != nil {
			t.Fatalf("Error terminando contenedor de Localstack: %v", err)
		}
	}()

	s3Client, err := setupS3Client(ctx, localstack.optsFunc)
	require.NoError(t, err, "Error creando cliente S3")

	err = createBucket(ctx, s3Client, testBucket)
	require.NoError(t, err, "Error creando bucket de prueba")

	testConfig := &config.Config{
		AWS: config.AWSConfig{
			Region: region,
		},
		Storage: config.StorageConfig{
			S3Config: config.S3Config{
				BucketName: testBucket,
			},
		},
	}

	logger, err := logging.NewDevelopmentLogger()
	require.NoError(t, err)

	storage := &S3Storage{
		client: s3Client,
		config: testConfig,
		logger: logger,
	}

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
		defer func() {
			if err := reader.Close(); err != nil {
				t.Fatalf("Error cerrando lector: %v", err)
			}
		}()

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
		ctxCanceled, cancelFunc := context.WithCancel(ctx)
		cancelFunc()

		_, err := storage.GetAudio(ctxCanceled, "anyfile.mp3")
		require.ErrorIs(t, err, context.Canceled, "Debería detectar contexto cancelado")
	})
}

func bytesReader(data []byte) io.Reader {
	return bytes.NewReader(data)
}

func randomString(length int) string {
	b := make([]byte, length/2)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", b)
}
