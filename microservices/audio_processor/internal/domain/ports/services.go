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

	CoreService interface {
		ProcessMedia(ctx context.Context, media *model.Media, userID, requestID string) error
	}
)
