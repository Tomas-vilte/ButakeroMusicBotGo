package handler

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	operationID := c.Query("operation_id")
	songID := c.Query("song_id")

	if operationID == "" || songID == "" {
		c.Error(errors.ErrInvalidInput.WithMessage("faltan los parámetros 'operation_id' y/o 'song_id'"))
		return
	}

	if _, err := uuid.Parse(operationID); err != nil {
		c.Error(errors.ErrInvalidInput.WithMessage("operation_id inválido"))
		return
	}

	status, err := h.getOperationStatusUC.Execute(c.Request.Context(), operationID, songID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{
		"operation_id": operationID,
		"status":       status,
	})
}
