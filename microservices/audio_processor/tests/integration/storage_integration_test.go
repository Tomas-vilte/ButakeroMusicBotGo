package integration

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/storage/cloud"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestS3StorageIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Saltando test de integración en modo corto")
	}

	cfgApp := &config.Config{
		AWS: config.AWSConfig{
			Region: os.Getenv("AWS_REGION"),
		},
		Storage: config.StorageConfig{
			S3Config: &config.S3Config{
				BucketName: os.Getenv("BUCKET_NAME"),
			},
		},
	}

	// config
	if cfgApp.Storage.S3Config.BucketName == "" || cfgApp.AWS.Region == "" {
		t.Fatal("BUCKET_NAME y REGION deben estar configurados para los tests de integración")
	}

	s3Storage, err := cloud.NewS3Storage(cfgApp)
	if err != nil {
		t.Fatalf("Error al crear S3Storage: %v", err)
	}

	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(), awsCfg.WithRegion(cfgApp.AWS.Region))
	if err != nil {
		t.Fatalf("Error al cargar la configuración de AWS: %v", err)
	}
	s3Client := s3.NewFromConfig(cfg)

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
			Bucket: aws.String(cfgApp.Storage.S3Config.BucketName),
			Key:    aws.String("audio/" + fileName),
		}
		result, err := s3Client.GetObject(context.Background(), getObjectInput)
		if err != nil {
			t.Fatalf("Error al obtener el objeto de S3: %v", err)
		}
		defer result.Body.Close()

		downloadedContent, err := io.ReadAll(result.Body)
		if err != nil {
			t.Fatalf("Error al leer el contenido del objeto: %v", err)
		}

		if string(downloadedContent) != content {
			t.Errorf("El contenido descargado no coincide. Obtenido: %s, Esperado: %s", string(downloadedContent), content)
		}

		// clear
		deleteObjectInput := &s3.DeleteObjectInput{
			Bucket: aws.String(cfgApp.Storage.S3Config.BucketName),
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

		expectedPublicURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", cfgApp.Storage.S3Config.BucketName, fileName)
		if fileData.PublicURL != expectedPublicURL {
			t.Errorf("PublicURL incorrecto. Obtenido: %s, Esperado: %s", fileData.PublicURL, expectedPublicURL)
		}

		// clear
		deleteObjectInput := &s3.DeleteObjectInput{
			Bucket: aws.String(cfgApp.Storage.S3Config.BucketName),
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
}
