package entity

type (
	MessageQueue struct {
		VideoID          string   `json:"video_id"`
		Message          string   `json:"message"`
		PlatformMetadata Metadata `json:"platform_metadata"`
		FileData         FileData `json:"file_data"`
		Success          bool     `json:"success"`
		Status           string   `json:"status"`
	}

	Metadata struct {
		Title        string `json:"title" bson:"title" dynamodbav:"title"`
		DurationMs   int64  `json:"duration_ms" bson:"duration_ms" dynamodbav:"duration_ms"`
		URL          string `json:"url" bson:"url" dynamodbav:"url"`
		ThumbnailURL string `json:"thumbnail_url" bson:"thumbnail_url" dynamodbav:"thumbnail_url"`
		Platform     string `json:"platform" bson:"platform" dynamodbav:"platform"`
	}

	FileData struct {
		FilePath string `json:"file_path" bson:"file_path" dynamodbav:"file_path"`
		FileSize string `json:"file_size" bson:"file_size" dynamodbav:"file_size"`
		FileType string `json:"file_type" bson:"file_type" dynamodbav:"file_type"`
	}
)
