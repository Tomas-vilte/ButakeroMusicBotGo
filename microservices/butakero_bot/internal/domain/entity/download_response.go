package entity

type (
	DownloadResponse struct {
		Provider string `json:"provider"`
		Status   string `json:"status"`
		Success  bool   `json:"success"`
		VideoID  string `json:"video_id"`
	}

	APIError struct {
		Error   ErrorDetail `json:"error"`
		Success bool        `json:"success"`
	}

	ErrorDetail struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
)
