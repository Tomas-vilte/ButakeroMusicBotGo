package config

import "time"

type (
	// Config es la estructura principal que agrupa todas las configuraciones
	Config struct {
		Environment string          `yaml:"environment"`
		Service     ServiceConfig   `yaml:"service"`
		AWS         *AWSConfig      `yaml:"aws,omitempty"`
		Messaging   MessagingConfig `yaml:"messaging"`
		Storage     StorageConfig   `yaml:"storage"`
		Database    DatabaseConfig  `yaml:"database"`
		API         APIConfig       `yaml:"api"`
	}

	// ServiceConfig contiene configuración general del servicio
	ServiceConfig struct {
		MaxAttempts int           `yaml:"max_attempts"`
		Timeout     time.Duration `yaml:"timeout"`
	}

	// AWSConfig contiene todas las configuraciones relacionadas con AWS
	AWSConfig struct {
		Region      string            `yaml:"region"`
		Credentials CredentialsConfig `yaml:"credentials"`
	}

	// CredentialsConfig maneja las credenciales de acceso
	CredentialsConfig struct {
		AccessKey string `yaml:"access_key"`
		SecretKey string `yaml:"secret_key"`
	}

	// MessagingConfig maneja la configuración de mensajería
	MessagingConfig struct {
		Type  string       `yaml:"type"` // kafka o sqs
		Kafka *KafkaConfig `yaml:"kafka,omitempty"`
		SQS   *SQSConfig   `yaml:"sqs,omitempty"`
	}

	// KafkaConfig configuración específica de Kafka
	KafkaConfig struct {
		Brokers []string `yaml:"brokers"`
		Topic   string   `yaml:"topic"`
	}

	// SQSConfig configuración específica de SQS
	SQSConfig struct {
		QueueURL string `yaml:"queue_url"`
		Topic    string `yaml:"topic"`
	}

	// StorageConfig maneja la configuración de almacenamiento
	StorageConfig struct {
		Type     string    `yaml:"type"` // s3 o local
		S3Config *S3Config `yaml:"s3,omitempty"`
	}

	// S3Config configuración específica de S3
	S3Config struct {
		BucketName string `yaml:"bucket_name"`
	}

	// DatabaseConfig maneja la configuración de base de datos
	DatabaseConfig struct {
		Type     string          `yaml:"type"` // mongodb o dynamodb
		Mongo    *MongoConfig    `yaml:"mongodb,omitempty"`
		DynamoDB *DynamoDBConfig `yaml:"dynamodb,omitempty"`
	}

	// MongoConfig configuración específica de MongoDB
	MongoConfig struct {
		User        string      `yaml:"user"`
		Password    string      `yaml:"password"`
		Port        string      `yaml:"port"`
		Host        string      `yaml:"host"`
		Database    string      `yaml:"database"`
		Collections Collections `yaml:"collections"`
	}

	// DynamoDBConfig configuración específica de DynamoDB
	DynamoDBConfig struct {
		Tables Tables `yaml:"tables"`
	}

	// Collections nombres de colecciones para MongoDB
	Collections struct {
		Songs      string `yaml:"songs"`
		Operations string `yaml:"operations"`
	}

	// Tables nombres de tablas para DynamoDB
	Tables struct {
		Songs      string `yaml:"songs"`
		Operations string `yaml:"operations"`
	}

	// APIConfig maneja la configuración de APIs externas
	APIConfig struct {
		YouTube YouTubeConfig `yaml:"youtube"`
		OAuth2  OAuth2Config  `yaml:"oauth2"`
	}

	// YouTubeConfig configuración específica de YouTube API
	YouTubeConfig struct {
		ApiKey string `yaml:"api_key"`
	}

	// OAuth2Config configuración de OAuth2
	OAuth2Config struct {
		Enabled string `yaml:"enabled"`
	}
)
