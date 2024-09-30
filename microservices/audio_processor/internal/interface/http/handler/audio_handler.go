package handler

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/usecase"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AudioHandler struct {
	initiateDownloadUC   usecase.InitialDownloadUseCase
	getOperationStatusUC usecase.GetOperationStatusUseCase
}

func NewAudioHandler(initiateDownloadUC usecase.InitialDownloadUseCase, getOperationStatusUC usecase.GetOperationStatusUseCase) *AudioHandler {
	return &AudioHandler{
		initiateDownloadUC:   initiateDownloadUC,
		getOperationStatusUC: getOperationStatusUC,
	}
}

func (h *AudioHandler) InitiateDownload(c *gin.Context) {
	song := c.Query("song")
	if song == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Falta el parametro 'song'",
		})
		return
	}

	operationID, err := h.initiateDownloadUC.Execute(c.Request.Context(), song)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"operation_id": operationID})
}

func (h *AudioHandler) GetOperationStatus(c *gin.Context) {
	operationID := c.Query("operation_id")
	songID := c.Query("song_id")

	if operationID == "" || songID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Faltan los parámetros 'operationID' y/o 'songID'",
		})
		return
	}
	status, err := h.getOperationStatusUC.Execute(c.Request.Context(), operationID, songID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}
