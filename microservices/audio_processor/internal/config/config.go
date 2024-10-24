package config

import "time"

type Config struct {
	MaxAttempts           int
	Timeout               time.Duration
	TimeoutSQS            time.Duration
	BucketName            string
	Region                string
	OperationResultsTable string
	SongsTable            string
	YouTubeApiKey         string
	AccessKey             string
	SecretKey             string
	Environment           string
	QueueURL              string
	Brokers               []string
	Topic                 string
	MongoUser             string
	MongoPassword         string
	MongoPort             string
	MongoHost             string
	MongoDB               string
	SongsCollection       string
}
