package entity

type DownloadResponse struct {
	OperationID string `json:"operation_id"`
	SongID      string `json:"song_id"`
	Status      string `json:"status"`
	Provider    string `json:"provider"`
	CreatedAt   string `json:"created_at"`
}
