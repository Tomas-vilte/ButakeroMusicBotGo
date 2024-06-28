package s3uploader

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"testing"
)

func TestS3Uploader_UploadFile(t *testing.T) {
	// Configurar el mock
	mockSvc := new(MockS3API)
	uploader := NewS3UploaderWithClient(mockSvc)

	// Preparar datos de prueba
	ctx := context.TODO()
	filePath := "./test.txt"
	bucketName := "test-bucket"
	key := "test-key"

	// Mock de la llamada al cliente S3API
	mockSvc.On("PutObjectWithContext", ctx, mock.AnythingOfType("*s3.PutObjectInput")).Return(&s3.PutObjectOutput{}, nil)

	// Ejecutar el método bajo prueba
	err := uploader.UploadFile(ctx, filePath, bucketName, key)

	// Verificar que no haya error
	assert.NoError(t, err, "Se esperaba que no hubiera error al subir archivo")

	// Verificar que el método del mock fue llamado correctamente
	mockSvc.AssertCalled(t, "PutObjectWithContext", ctx, mock.AnythingOfType("*s3.PutObjectInput"))
}

func TestS3Uploader_UploadContent(t *testing.T) {
	// Configurar el mock
	mockSvc := new(MockS3API)
	uploader := NewS3UploaderWithClient(mockSvc)

	// Preparar datos de prueba
	ctx := context.TODO()
	content := strings.NewReader("test content")
	bucketName := "test-bucket"
	key := "test-key"

	// Mock de la llamada al cliente S3API
	mockSvc.On("PutObjectWithContext", ctx, mock.AnythingOfType("*s3.PutObjectInput")).Return(&s3.PutObjectOutput{}, nil)

	// Ejecutar el método bajo prueba
	err := uploader.UploadContent(ctx, content, bucketName, key)

	// Verificar que no haya error
	assert.NoError(t, err, "Se esperaba que no hubiera error al subir contenido")

	// Verificar que el método del mock fue llamado correctamente
	mockSvc.AssertCalled(t, "PutObjectWithContext", ctx, mock.AnythingOfType("*s3.PutObjectInput"))
}

func TestS3Uploader_UploadFile_ErrorOpeningFile(t *testing.T) {
	// Configurar el mock
	mockSvc := new(MockS3API)
	uploader := NewS3UploaderWithClient(mockSvc)

	// Preparar datos de prueba
	ctx := context.TODO()
	filePath := "./noexisting.txt"
	bucketName := "test-bucket"
	key := "test-key"

	// Ejecutar el método bajo prueba
	err := uploader.UploadFile(ctx, filePath, bucketName, key)

	// Verificar que se produzca un error al abrir el archivo
	assert.Error(t, err, "Se esperaba un error al abrir el archivo")
	assert.Contains(t, err.Error(), "error al abrir el archivo")
}

func TestS3Uploader_UploadContent_ErrorPuttingObject(t *testing.T) {
	// Configurar el mock
	mockSvc := new(MockS3API)
	uploader := NewS3UploaderWithClient(mockSvc)

	// Preparar datos de prueba
	ctx := context.TODO()
	content := strings.NewReader("test content")
	bucketName := "test-bucket"
	key := "test-key"

	// Mock de la llamada al cliente S3API para retornar un error
	mockError := fmt.Errorf("error simulado al subir objeto a S3")
	mockSvc.On("PutObjectWithContext", ctx, mock.AnythingOfType("*s3.PutObjectInput")).Return(&s3.PutObjectOutput{}, mockError)

	// Ejecutar el método bajo prueba
	err := uploader.UploadContent(ctx, content, bucketName, key)

	// Verificar que se produzca un error al subir el contenido
	assert.Error(t, err, "Se esperaba un error al subir contenido a S3")
	assert.Contains(t, err.Error(), "error simulado al subir objeto a S3")
}
