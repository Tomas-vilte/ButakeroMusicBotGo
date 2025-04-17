package queue

type (
	DownloadStatusMessage struct {
		RequestID        string       `json:"request_id"`
		UserID           string       `json:"user_id"`
		VideoID          string       `json:"video_id"`
		Message          string       `json:"message"`
		PlatformMetadata SongMetadata `json:"platform_metadata"`
		FileData         FileData     `json:"file_data"`
		Success          bool         `json:"success"`
		Status           string       `json:"status"`
	}

	SongMetadata struct {
		Title        string `json:"title"`
		DurationMs   int64  `json:"duration_ms"`
		URL          string `json:"url"`
		ThumbnailURL string `json:"thumbnail_url"`
		Platform     string `json:"platform"`
	}

	FileData struct {
		FilePath string `json:"file_path"`
		FileSize string `json:"file_size"`
		FileType string `json:"file_type"`
	}
)
