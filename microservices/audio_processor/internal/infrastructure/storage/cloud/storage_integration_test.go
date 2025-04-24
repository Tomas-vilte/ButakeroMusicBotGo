//go:build integration

package cloud

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"strings"
	"testing"
	"time"
)

const (
	testBucket  = "test-bucket"
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

	cfg, err := awsCfg.LoadDefaultConfig(ctx,
		awsCfg.WithRegion(region),
		awsCfg.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
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
			S3Config: &config.S3Config{
				BucketName: testBucket,
			},
		},
	}

	log, err := logger.NewDevelopmentLogger()
	require.NoError(t, err)

	// config
	if testConfig.Storage.S3Config.BucketName == "" || testConfig.AWS.Region == "" {
		t.Fatal("BUCKET_NAME y REGION deben estar configurados para los tests de integración")
	}

	s3Storage := S3Storage{
		log:    log,
		Client: s3Client,
		Config: testConfig,
	}

	t.Run("Submit and verify file", func(t *testing.T) {
		// arrange
		fileName := fmt.Sprintf("test-file-%d.txt", time.Now().UnixNano())
		content := "Este es un archivo de prueba para la integración"

		// act
		err := s3Storage.UploadFile(context.Background(), fileName, strings.NewReader(content))
		if err != nil {
			t.Fatalf("Error al subir el archivo: %v", err)
		}

		// assert
		getObjectInput := &s3.GetObjectInput{
			Bucket: aws.String(testConfig.Storage.S3Config.BucketName),
			Key:    aws.String("audio/" + fileName),
		}
		result, err := s3Client.GetObject(context.Background(), getObjectInput)
		if err != nil {
			t.Fatalf("Error al obtener el objeto de S3: %v", err)
		}
		defer func() {
			if err := result.Body.Close(); err != nil {
				t.Fatalf("Error al cerrar el cuerpo del objeto: %v", err)
			}
		}()

		downloadedContent, err := io.ReadAll(result.Body)
		if err != nil {
			t.Fatalf("Error al leer el contenido del objeto: %v", err)
		}

		if string(downloadedContent) != content {
			t.Errorf("El contenido descargado no coincide. Obtenido: %s, Esperado: %s", string(downloadedContent), content)
		}

		// clear
		deleteObjectInput := &s3.DeleteObjectInput{
			Bucket: aws.String(testConfig.Storage.S3Config.BucketName),
			Key:    aws.String("audio/" + fileName),
		}

		_, err = s3Client.DeleteObject(context.Background(), deleteObjectInput)
		if err != nil {
			t.Fatalf("Error al eliminar el objeto de prueba: %v", err)
		}
	})

	t.Run("Intent submit file which body null", func(t *testing.T) {
		// act
		err := s3Storage.UploadFile(context.Background(), "test-null-body.txt", nil)

		// assert
		if err == nil {
			t.Fatal("Se esperaba un error al subir un archivo con cuerpo nulo, pero no se obtuvo ninguno")
		}

		if err.Error() != "el cuerpo no puede ser nulo" {
			t.Errorf("Mensaje de error inesperado. Obtenido: %s, Esperado: 'el cuerpo no puede ser nulo'", err.Error())
		}
	})

	t.Run("Get file metadata for existing file", func(t *testing.T) {
		// arrange
		fileName := fmt.Sprintf("test-metadata-%d.dca", time.Now().UnixNano())
		content := "Este es un archivo de prueba"

		// subimos archivo
		err := s3Storage.UploadFile(context.Background(), fileName, strings.NewReader(content))
		if err != nil {
			t.Fatalf("Error al subir el archivo para la prueba de metadata: %v", err)
		}

		// act
		fileData, err := s3Storage.GetFileMetadata(context.Background(), fileName)

		// assert
		if err != nil {
			t.Fatalf("Error al obtener metadata del archivo: %v", err)
		}

		if fileData == nil {
			t.Fatal("FileData es nil, se esperaba un objeto no nulo")
		}
		expectedFilePath := fmt.Sprintf("audio/%s", fileName)
		if fileData.FilePath != expectedFilePath {
			t.Errorf("FilePath incorrecto. Obtenido: %s, Esperado: %s", fileData.FilePath, expectedFilePath)
		}

		// Verificar FileType (asumiendo que se establece correctamente al subir)
		if fileData.FileType != "application/octet-stream" {
			t.Errorf("FileType incorrecto. Obtenido: %s, Esperado: audio/mpeg", fileData.FileType)
		}

		// Verificar FileSize (el formato exacto dependerá de tu implementación)
		if fileData.FileSize == "" {
			t.Error("FileSize está vacío, se esperaba un valor")
		}
		// clear
		deleteObjectInput := &s3.DeleteObjectInput{
			Bucket: aws.String(testConfig.Storage.S3Config.BucketName),
			Key:    aws.String("audio/" + fileName),
		}

		_, err = s3Client.DeleteObject(context.Background(), deleteObjectInput)
		if err != nil {
			t.Fatalf("Error al eliminar el objeto de prueba de metadata: %v", err)
		}
	})

	t.Run("Get file metadata for non-existent file", func(t *testing.T) {
		// act
		nonExistentFileName := "non-existent-file.mp3"
		fileData, err := s3Storage.GetFileMetadata(context.Background(), nonExistentFileName)

		// assert
		if err == nil {
			t.Fatal("Se esperaba un error al obtener metadata de un archivo no existente, pero no se obtuvo ninguno")
		}

		if fileData != nil {
			t.Error("Se esperaba que FileData fuera nil para un archivo no existente")
		}

		if !strings.Contains(err.Error(), "error obteniendo metadata del archivo de S3") {
			t.Errorf("Mensaje de error inesperado. Obtenido: %s, Esperado que contenga: 'error obteniendo metadata del archivo de S3'", err.Error())
		}
	})

	t.Run("Get file content for existing file", func(t *testing.T) {
		// arrange
		fileName := fmt.Sprintf("test-content-%d.txt", time.Now().UnixNano())
		content := "Este es un archivo de prueba para obtener el contenido"

		err := s3Storage.UploadFile(context.Background(), fileName, strings.NewReader(content))
		if err != nil {
			t.Fatalf("Error al subir el archivo: %v", err)
		}

		// act
		fileContent, err := s3Storage.GetFileContent(context.Background(), "audio/", fileName)
		if err != nil {
			t.Fatalf("Error al obtener el contenido del archivo: %v", err)
		}
		defer func() {
			if err := fileContent.Close(); err != nil {
				t.Fatalf("Error al cerrar el lector: %v", err)
			}
		}()

		downloadedContent, err := io.ReadAll(fileContent)
		if err != nil {
			t.Fatalf("Error al leer el contenido del archivo: %v", err)
		}

		// assert
		if string(downloadedContent) != content {
			t.Errorf("El contenido descargado no coincide. Obtenido: %s, Esperado: %s", string(downloadedContent), content)
		}

		// clear
		deleteObjectInput := &s3.DeleteObjectInput{
			Bucket: aws.String(testConfig.Storage.S3Config.BucketName),
			Key:    aws.String("audio/" + fileName),
		}

		_, err = s3Client.DeleteObject(context.Background(), deleteObjectInput)
		if err != nil {
			t.Fatalf("Error al eliminar el archivo: %v", err)
		}
	})
}
