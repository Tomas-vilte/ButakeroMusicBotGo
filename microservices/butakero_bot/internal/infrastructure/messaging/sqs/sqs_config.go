package sqs

type SQSConfig struct {
	QueueURL        string
	MaxMessages     int32
	WaitTimeSeconds int32
}
