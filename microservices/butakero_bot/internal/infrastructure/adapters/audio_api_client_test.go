//go:build !integration

package adapters

import (
	"context"
	"encoding/json"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNewAudioAPIClient_ValidConfig(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	config := AudioAPIClientConfig{
		BaseURL:         "https://valid-url.com",
		Timeout:         5 * time.Second,
		MaxIdleConns:    10,
		MaxConnsPerHost: 20,
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	client, err := NewAudioAPIClient(config, mockLogger)

	require.NoError(t, err)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, 5*time.Second, client.httpClient.Timeout)
}

func TestNewAudioAPIClient_InvalidBaseURL(t *testing.T) {
	// Arrange
	loggerMock := new(logging.MockLogger)
	invalidConfig := AudioAPIClientConfig{
		BaseURL: ":invalid-url",
	}

	// Act
	client, err := NewAudioAPIClient(invalidConfig, loggerMock)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "La URL base no es válida")
	assert.Nil(t, client)
}

func TestDownloadSong_EmptySongName(t *testing.T) {
	loggerMock := new(logging.MockLogger)
	client := &AudioAPIClient{logger: loggerMock}

	loggerMock.On("With", mock.Anything, mock.Anything).Return(loggerMock)
	loggerMock.On("Error", mock.Anything, mock.Anything)
	_, err := client.DownloadSong(context.Background(), "", "youtube")

	assert.Contains(t, err.Error(), "Faltan algunos parámetros provider_type/song")
}

func TestDownloadSong_HTTPError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer testServer.Close()

	mockLogger := new(logging.MockLogger)

	baseURL, _ := url.Parse(testServer.URL)
	client := &AudioAPIClient{
		baseURL:    baseURL,
		logger:     mockLogger,
		httpClient: &http.Client{Timeout: time.Nanosecond},
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	_, err := client.DownloadSong(context.Background(), "test-song", "youtube")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "El servicio de música no está disponible en este momento")
}

func TestDownloadSong_NonOKStatus(t *testing.T) {
	// Arrange
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServer.Close()

	mockLogger := new(logging.MockLogger)

	baseURL, _ := url.Parse(testServer.URL)
	client := &AudioAPIClient{
		baseURL:    baseURL,
		logger:     mockLogger,
		httpClient: &http.Client{},
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	// Act
	_, err := client.DownloadSong(context.Background(), "test-song", "youtube")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error en la solicitud (Código: 500)")
}

func TestDownloadSong_APIError_DuplicateRecord(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(entity.APIError{
			Error: entity.ErrorDetail{
				Code:    "duplicate_record",
				Message: "El video con ID 'SRXH9AbT280' ya esta registado",
			},
		})
	}))
	defer testServer.Close()

	mockLogger := new(logging.MockLogger)

	baseURL, _ := url.Parse(testServer.URL)
	client := &AudioAPIClient{
		baseURL:    baseURL,
		logger:     mockLogger,
		httpClient: &http.Client{},
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	_, err := client.DownloadSong(context.Background(), "test-song", "youtube")

	require.Error(t, err)
	assert.Equal(t, err.Error(), "Error al procesar el mensaje de error")
}

func TestDownloadSong_APIError_ProviderNotFound(t *testing.T) {
	// Arrange
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(entity.APIError{
			Error: entity.ErrorDetail{
				Code:    "provider_not_found",
				Message: "El proveedor no fue encontrado",
			},
		})
	}))
	defer testServer.Close()

	mockLogger := new(logging.MockLogger)

	baseURL, _ := url.Parse(testServer.URL)
	client := &AudioAPIClient{
		baseURL:    baseURL,
		logger:     mockLogger,
		httpClient: &http.Client{},
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	// Act
	_, err := client.DownloadSong(context.Background(), "test-song", "youtube")

	// Assert
	require.Error(t, err)
	assert.Equal(t, "El proveedor de música no fue encontrado", err.Error())
}

func TestDownloadSong_Success(t *testing.T) {
	// Arrange
	expectedResponse := entity.DownloadResponse{
		VideoID: "123",
		Status:  "processing",
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer testServer.Close()

	baseURL, _ := url.Parse(testServer.URL)
	loggerMock := new(logging.MockLogger)
	loggerMock.On("With", mock.Anything, mock.Anything).Return(loggerMock)
	loggerMock.On("Info", mock.Anything, mock.Anything).Return()
	loggerMock.On("Debug", mock.Anything, mock.Anything).Return()

	client := &AudioAPIClient{
		baseURL:    baseURL,
		logger:     loggerMock,
		httpClient: &http.Client{},
	}

	// Act
	resp, err := client.DownloadSong(context.Background(), "test-song", "youtube")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedResponse.VideoID, resp.VideoID)
	assert.Equal(t, expectedResponse.Status, resp.Status)
	loggerMock.AssertExpectations(t)
}
