package entity

import "time"

type (
	StatusMessage struct {
		Status Status `json:"status"`
	}

	Status struct {
		ID             string    `json:"id"`
		SK             string    `json:"sk"`
		Status         string    `json:"status"`
		Message        string    `json:"message"`
		Metadata       Metadata  `json:"metadata"`
		FileData       FileData  `json:"file_data"`
		ProcessingDate time.Time `json:"processing_date"`
		Success        bool      `json:"success"`
		Attempts       int       `json:"attempts"`
		Failures       int       `json:"failures"`
	}

	Metadata struct {
		ID           string `bson:"_id" dynamodbav:"_id"`
		VideoID      string `bson:"video_id" dynamodbav:"video_id"`
		Title        string `bson:"title" dynamodbav:"title"`
		DurationMs   int64  `bson:"duration_ms" dynamodbav:"duration_ms"`
		URL          string `bson:"url" dynamodbav:"url"`
		ThumbnailURL string `bson:"thumbnail_url" dynamodbav:"thumbnail_url"`
		Platform     string `bson:"platform" dynamodbav:"platform"`
	}

	FileData struct {
		FilePath string `bson:"file_path" dynamodbav:"file_path"`
		FileSize string `bson:"file_size" dynamodbav:"file_size"`
		FileType string `bson:"file_type" dynamodbav:"file_type"`
	}
)
