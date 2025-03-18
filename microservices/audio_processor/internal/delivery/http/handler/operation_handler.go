package handler

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/usecase"
	"github.com/gin-gonic/gin"
)

type OperationHandler struct {
	getOperationStatusUC usecase.GetOperationStatusUseCase
}

func NewOperationHandler(getOperationStatusUC usecase.GetOperationStatusUseCase) *OperationHandler {
	return &OperationHandler{
		getOperationStatusUC: getOperationStatusUC,
	}
}

func (h *OperationHandler) GetOperationStatus(c *gin.Context) {
	videoID := c.Query("video_id")

	if videoID == "" {
		_ = c.Error(errors.ErrInvalidInput.WithMessage("faltan los parámetros 'song_id'"))
		return
	}

	if !isValidSongID(videoID) {
		_ = c.Error(errors.ErrInvalidInput.WithMessage("video_id inválido"))
		return
	}

	response, err := h.getOperationStatusUC.Execute(c.Request.Context(), videoID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(200, response)
}

func isValidSongID(songID string) bool {
	return len(songID) > 0 && len(songID) <= 100
}
