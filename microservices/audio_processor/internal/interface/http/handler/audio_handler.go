package handler

import (
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/usecase"
	"github.com/gin-gonic/gin"
)

type AudioHandler struct {
	initiateDownloadUC usecase.InitialDownloadUseCase
}

func NewAudioHandler(initiateDownloadUC usecase.InitialDownloadUseCase) *AudioHandler {
	return &AudioHandler{
		initiateDownloadUC: initiateDownloadUC,
	}
}

func (h *AudioHandler) InitiateDownload(c *gin.Context) {
	song := c.Query("song")
	providerType := c.Query("provider_type")

	if song == "" || providerType == "" {
		missingParams := make([]string, 0)
		if song == "" {
			missingParams = append(missingParams, "song")
		}
		if providerType == "" {
			missingParams = append(missingParams, "provider_type")
		}

		c.Error(errors.ErrInvalidInput.WithMessage(
			fmt.Sprintf("faltan parámetros requeridos: %v", missingParams),
		))
		return
	}

	result, err := h.initiateDownloadUC.Execute(c.Request.Context(), song, providerType)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{
		"operation_id": result.ID,
		"song_id":      result.SongID,
		"provider":     providerType,
		"status":       result.Status,
		"created_at":   result.CreatedAt,
	})
}
