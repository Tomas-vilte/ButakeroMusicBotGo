package controller

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type MediaController struct {
	mediaRepository ports.MediaRepository
}

func NewMediaController(mediaRepository ports.MediaRepository) *MediaController {
	return &MediaController{mediaRepository: mediaRepository}
}

func (mc *MediaController) GetMediaByID(c *gin.Context) {
	videoID := c.Query("video_id")

	if videoID == "" {
		_ = c.Error(errors.ErrInvalidInput.WithMessage("falta el parametro 'video_id'"))
		return
	}

	if !isValidSongID(videoID) {
		_ = c.Error(errors.ErrInvalidInput.WithMessage("video_id inválido"))
		return
	}

	media, err := mc.mediaRepository.GetMediaByID(c.Request.Context(), videoID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    media,
		"success": true,
	})
}

func (mc *MediaController) SearchMediaByTitle(c *gin.Context) {
	title := c.Query("title")
	if len(title) < 3 {
		_ = c.Error(errors.ErrInvalidInput.WithMessage("el título debe tener al menos 3 caracteres"))
		return
	}

	medias, err := mc.mediaRepository.GetMediaByTitle(c.Request.Context(), title)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    medias,
		"success": true,
	})
}

func isValidSongID(songID string) bool {
	return len(songID) > 0 && len(songID) <= 100
}
