package config

import (
	"strconv"
	"time"
)

type (
	Config struct {
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
		OAuth2                string
		Mongo                 MongoConfig
	}

	MongoConfig struct {
		User                       string
		Password                   string
		Port                       string
		Host                       string
		Database                   string
		SongsCollection            string
		OperationResultsCollection string
	}
)

func (c *Config) ParseBool() bool {
	enabled, err := strconv.ParseBool(c.OAuth2)
	if err != nil {
		return false
	}
	return enabled
}
