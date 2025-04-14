package model

type (
	MediaProcessingMessage struct {
		UserID           string            `json:"user_id"`
		RequestID        string            `json:"request_id"`
		VideoID          string            `json:"video_id"`
		FileData         *FileData         `json:"file_data"`
		PlatformMetadata *PlatformMetadata `json:"platform_metadata"`
		ReceiptHandle    string            `json:"receipt_handle,omitempty"`
		Message          string            `json:"message"`
		Success          bool              `json:"success"`
		Status           string            `json:"status"`
	}
)
