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
		ID         string `json:"id"`
		VideoID    string `json:"video_id"`
		Title      string `json:"title"`
		Duration   string `json:"duration"`
		URLYoutube string `json:"url_youtube"`
		Thumbnail  string `json:"thumbnail"`
		Platform   string `json:"platform"`
	}

	FileData struct {
		FilePath  string `json:"file_path"`
		FileSize  string `json:"file_size"`
		FileType  string `json:"file_type"`
		PublicURL string `json:"public_url"`
	}
)
