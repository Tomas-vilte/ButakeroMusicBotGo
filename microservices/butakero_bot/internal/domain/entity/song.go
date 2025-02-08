package entity

type Song struct {
	ID          string `bson:"_id,omitempty"`
	VideoID     string `bson:"video_id"`
	Title       string `bson:"title"`
	Duration    string `bson:"duration"`
	URLYoutube  string `bson:"url_youtube"`
	Thumbnail   string `bson:"thumbnail"`
	Platform    string `bson:"platform"`
	FilePath    string `bson:"file_path"`
	FileSize    string `bson:"file_size"`
	FileType    string `bson:"file_type"`
	PublicURL   string `bson:"public_url"`
	ProcessDate string `bson:"processing_date"`
	Success     bool   `bson:"success"`
	Attempts    int    `bson:"attempts"`
	Failures    int    `bson:"failures"`
}
