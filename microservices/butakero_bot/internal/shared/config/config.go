package config

type Config struct {
	Storage       StorageConfig
	AWS           AWSConfig
	CommandPrefix string
	Discord       Discord
	Kafka         Kafka
}

type Discord struct {
	Token string
}

type Kafka struct {
	Brokers []string
	Topic   string
}

type StorageConfig struct {
	S3Config    S3Config
	LocalConfig LocalConfig
}

type S3Config struct {
	BucketName string
	Region     string
}

type LocalConfig struct {
	Directory string
}

type AWSConfig struct {
	Region string
}
