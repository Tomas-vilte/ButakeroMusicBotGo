package model

type (
	MediaProcessingMessage struct {
		VideoID          string            `json:"video_id"`
		FileData         *FileData         `json:"file_data"`
		PlatformMetadata *PlatformMetadata `json:"platform_metadata"`
		ReceiptHandle    string            `json:"receipt_handle,omitempty"`
		Status           string            `json:"status"`
		Message          string            `json:"message"`
	}
)
