package ports

import (
	"bytes"
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
)

type (
	VideoService interface {
		GetMediaDetails(ctx context.Context, input string, providerType string) (*model.MediaDetails, error)
	}

	AudioDownloadService interface {
		DownloadAndEncode(ctx context.Context, url string) (*bytes.Buffer, error)
	}

	AudioStorageService interface {
		StoreAudio(ctx context.Context, buffer *bytes.Buffer, songName string) (*model.FileData, error)
	}

	TopicPublisherService interface {
		PublishMediaProcessed(ctx context.Context, message *model.MediaProcessingMessage) error
	}

	MediaService interface {
		CreateMedia(ctx context.Context, media *model.Media) error
		GetMediaByID(ctx context.Context, videoID string) (*model.Media, error)
		GetMediaByTitle(ctx context.Context, title string) ([]*model.Media, error)
		UpdateMedia(ctx context.Context, videoID string, status *model.Media) error
		DeleteMedia(ctx context.Context, videoID string) error
	}

	CoreService interface {
		ProcessMedia(ctx context.Context, mediaDetails *model.MediaDetails, userID, requestID string) error
	}

	OperationService interface {
		StartOperation(ctx context.Context, mediaDetails *model.MediaDetails) (*model.OperationInitResult, error)
	}
)
