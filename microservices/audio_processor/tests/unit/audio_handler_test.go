package unit

import (
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/interface/http/handler"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAudioHandler_InitiateDownload(t *testing.T) {
	t.Run("Successful download initiation", func(t *testing.T) {
		mockInitialDownloadUC := new(MockInitiateDownloadUC)
		mockGetOperationStatusUC := new(MockGetOperationStatusUC)

		handlerHttp := handler.NewAudioHandler(mockInitialDownloadUC, mockGetOperationStatusUC)

		mockInitialDownloadUC.On("Execute", mock.Anything, "test_song").Return("operation_123", "song_id", nil)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/download?song=test_song", nil)

		handlerHttp.InitiateDownload(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "operation_123")
		assert.Contains(t, w.Body.String(), "song_id")
		mockInitialDownloadUC.AssertExpectations(t)
	})

	t.Run("Missing song parameter", func(t *testing.T) {
		mockInitialDownloadUC := new(MockInitiateDownloadUC)
		mockGetOperationStatusUC := new(MockGetOperationStatusUC)

		handlerHttp := handler.NewAudioHandler(mockInitialDownloadUC, mockGetOperationStatusUC)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/download", nil)

		handlerHttp.InitiateDownload(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Falta el parametro 'song'")

	})

	t.Run("Use case error", func(t *testing.T) {
		mockInitialDownloadUC := new(MockInitiateDownloadUC)
		mockGetOperationStatusUC := new(MockGetOperationStatusUC)

		mockInitialDownloadUC.On("Execute", mock.Anything, "test_song").Return("", "", errors.New("use case error"))

		handlerHttp := handler.NewAudioHandler(mockInitialDownloadUC, mockGetOperationStatusUC)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/download?song=test_song", nil)

		handlerHttp.InitiateDownload(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "use case error")
		mockInitialDownloadUC.AssertExpectations(t)
	})
}

func TestAudioHandler_GetOperationStatus(t *testing.T) {
	t.Run("Successful status retrieval", func(t *testing.T) {
		mockInitialDownloadUC := new(MockInitiateDownloadUC)
		mockGetOperationStatusUC := new(MockGetOperationStatusUC)

		handlerHttp := handler.NewAudioHandler(mockInitialDownloadUC, mockGetOperationStatusUC)

		expectedResult := &model.OperationResult{
			Status: "completed",
		}

		mockGetOperationStatusUC.On("Execute", mock.Anything, "operation_123", "song_123").Return(expectedResult, nil)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/status?operation_id=operation_123&song_id=song_123", nil)

		handlerHttp.GetOperationStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "completed")
		mockGetOperationStatusUC.AssertExpectations(t)
	})

	t.Run("Missing parameters", func(t *testing.T) {
		mockInitialDownloadUC := new(MockInitiateDownloadUC)
		mockGetOperationStatusUC := new(MockGetOperationStatusUC)

		handlerHttp := handler.NewAudioHandler(mockInitialDownloadUC, mockGetOperationStatusUC)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/status", nil)

		handlerHttp.GetOperationStatus(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Faltan los parámetros 'operationID' y/o 'songID'")
	})

	t.Run("Use case error", func(t *testing.T) {
		mockInitialDownloadUC := new(MockInitiateDownloadUC)
		mockGetOperationStatusUC := new(MockGetOperationStatusUC)

		handlerHttp := handler.NewAudioHandler(mockInitialDownloadUC, mockGetOperationStatusUC)

		mockGetOperationStatusUC.On("Execute", mock.Anything, "operation_123", "song_123").Return(&model.OperationResult{}, errors.New("use case error"))

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/status?operation_id=operation_123&song_id=song_123", nil)

		handlerHttp.GetOperationStatus(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "use case error")
		mockGetOperationStatusUC.AssertExpectations(t)
	})
}
