package entity

import "time"

type (
	Song struct {
		ID          string `bson:"_id,omitempty" dynamodbav:"id"`
		VideoID     string `bson:"video_id" dynamodbav:"video_id"`
		Title       string `bson:"title" dynamodbav:"title"`
		Duration    string `bson:"duration" dynamodbav:"duration"`
		URLYoutube  string `bson:"url_youtube" dynamodbav:"url_youtube"`
		Thumbnail   string `bson:"thumbnail" dynamodbav:"thumbnail"`
		Platform    string `bson:"platform" dynamodbav:"platform"`
		FilePath    string `bson:"file_path" dynamodbav:"file_path"`
		FileSize    string `bson:"file_size" dynamodbav:"file_size"`
		FileType    string `bson:"file_type" dynamodbav:"file_type"`
		PublicURL   string `bson:"public_url" dynamodbav:"public_url"`
		ProcessDate string `bson:"processing_date" dynamodbav:"processing_date"`
		Success     bool   `bson:"success" dynamodbav:"success"`
		Attempts    int    `bson:"attempts" dynamodbav:"attempts"`
		Failures    int    `bson:"failures" dynamodbav:"failures"`
	}

	PlayedSong struct {
		Song
		Position time.Duration
	}
)
