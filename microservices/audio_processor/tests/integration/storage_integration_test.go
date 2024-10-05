package integration

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
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

	cfgApp := config.Config{
		BucketName: os.Getenv("BUCKET_NAME"),
		Region:     os.Getenv("REGION"),
		AccessKey:  os.Getenv("ACCESS_KEY"),
		SecretKey:  os.Getenv("SECRET_KEY"),
	}

	// config
	if cfgApp.BucketName == "" || cfgApp.Region == "" {
		t.Fatal("BUCKET_NAME y AWS_REGION deben estar configurados para los tests de integración")
	}

	s3Storage, err := storage.NewS3Storage(cfgApp)
	if err != nil {
		t.Fatalf("Error al crear S3Storage: %v", err)
	}

	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(), awsCfg.WithRegion(cfgApp.Region), awsCfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
		cfgApp.AccessKey, cfgApp.SecretKey, "")))
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
			Bucket: aws.String(cfgApp.BucketName),
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
			Bucket: aws.String(cfgApp.BucketName),
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
}
