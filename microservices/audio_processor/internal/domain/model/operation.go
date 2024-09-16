package model

type OperationResult struct {
	ID      string `json:"id"`
	SongID  string `json:"song_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    string `json:"data"`
}
