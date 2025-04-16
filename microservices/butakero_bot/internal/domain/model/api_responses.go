package model

import "time"

type MediaResponse struct {
	Data    *Media `json:"data"`
	Success bool   `json:"success"`
}

type MediaListResponse struct {
	Data    []*Media `json:"data"`
	Success bool     `json:"success"`
}

// ErrorResponse representa el formato estándar de errores
type ErrorResponse struct {
	Error   ErrorDetail `json:"error"`
	Success bool        `json:"success"`
}

// ErrorDetail contiene los detalles del error
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	VideoID string `json:"video_id,omitempty"`
}

type (
	Media struct {
		PK             string    `json:"-"`
		TitleLower     string    `json:"title_lower"`
		Status         string    `json:"status"`
		Message        string    `json:"message"`
		Metadata       Metadata  `json:"metadata"`
		FileData       FileData  `json:"file_data"`
		ProcessingDate time.Time `json:"processing_date"`
		Success        bool      `json:"success"`
		Attempts       int       `json:"attempts"`
		Failures       int       `json:"failures"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
		PlayCount      int       `json:"play_count"`
	}

	FileData struct {
		FilePath string `json:"file_path"`
		FileSize string `json:"file_size"`
		FileType string `json:"file_type"`
	}

	Metadata struct {
		Title        string `json:"title"`
		DurationMs   int64  `json:"duration_ms"`
		URL          string `json:"url"`
		ThumbnailURL string `json:"thumbnail_url"`
		Platform     string `json:"platform"`
	}
)
