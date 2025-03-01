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

	AudioProcessor interface {
		ProcessAudio(ctx context.Context, operationID string, metadata *model.MediaDetails) error
	}

	OperationStarter interface {
		StartOperation(ctx context.Context, songID string) (*model.OperationInitResult, error)
	}

	AudioDownloadService interface {
		DownloadAndEncode(ctx context.Context, url string) (*bytes.Buffer, error)
	}

	AudioStorageService interface {
		StoreAudio(ctx context.Context, buffer *bytes.Buffer, metadata *model.Metadata) (*model.FileData, error)
	}

	OperationsManager interface {
		HandleOperationSuccess(ctx context.Context, operationID string, metadata *model.Metadata, fileData *model.FileData) error
	}

	MessagingManager interface {
		SendProcessingMessage(ctx context.Context, operationID, status string, metadata *model.Metadata, attempts int) error
	}

	ErrorManagement interface {
		HandleProcessingError(ctx context.Context, operationID string, metadata *model.Metadata, stage string, attempts int, err error) error
	}
)
