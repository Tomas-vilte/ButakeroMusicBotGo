package entity

type DownloadResponse struct {
	Provider string `json:"provider"`
	Status   string `json:"status"`
	Success  bool   `json:"success"`
	VideoID  string `json:"video_id"`
}
