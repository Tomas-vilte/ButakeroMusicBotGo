package config

import "time"

type Config struct {
	MaxAttempts           int
	Timeout               time.Duration
	BucketName            string
	Region                string
	OperationResultsTable string
	SongsTable            string
	YouTubeApiKey         string
	AccessKey             string
	SecretKey             string
	Environment           string
}
