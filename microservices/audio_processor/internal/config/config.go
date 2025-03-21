package config

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/secretmanager"
	"os"
	"strconv"
	"time"
)

func LoadConfigLocal() *Config {
	return &Config{
		Environment: "local",
		Service: ServiceConfig{
			MaxAttempts: getEnvAsInt("SERVICE_MAX_ATTEMPTS", 1),
			Timeout:     time.Duration(getEnvAsInt("SERVICE_TIMEOUT", 1)) * time.Minute,
		},
		GinConfig: GinConfig{
			Mode: os.Getenv("GIN_MODE"),
		},
		Messaging: MessagingConfig{
			Type: "kafka",
			Kafka: &KafkaConfig{
				Brokers:   []string{os.Getenv("KAFKA_BROKERS")},
				Topic:     os.Getenv("KAFKA_TOPIC"),
				CaFile:    os.Getenv("KAFKA_CA_FILE"),
				CertFile:  os.Getenv("KAFKA_CERT_FILE"),
				KeyFile:   os.Getenv("KAFKA_KEY_FILE"),
				EnableTLS: os.Getenv("KAFKA_ENABLE_TLS") == "true",
			},
		},
		API: APIConfig{
			YouTube: YouTubeConfig{
				Cookies: os.Getenv("COOKIES_YOUTUBE"),
				ApiKey:  os.Getenv("YOUTUBE_API_KEY"),
			},
		},
		Storage: StorageConfig{
			Type: "local-storage",
			LocalConfig: &LocalConfig{
				BasePath: os.Getenv("LOCAL_STORAGE_PATH"),
			},
		},
		Database: DatabaseConfig{
			Type: "mongodb",
			Mongo: &MongoConfig{
				User:           os.Getenv("MONGO_USER"),
				Password:       os.Getenv("MONGO_PASSWORD"),
				Port:           os.Getenv("MONGO_PORT"),
				Host:           []string{os.Getenv("MONGO_HOST")},
				CaFile:         os.Getenv("MONGO_CA_FILE"),
				CertFile:       os.Getenv("MONGO_CERT_FILE"),
				KeyFile:        os.Getenv("MONGO_KEY_FILE"),
				Database:       os.Getenv("MONGO_DATABASE"),
				ReplicaSetName: os.Getenv("MONGO_REPLICA_SET_NAME"),
				EnableTLS:      os.Getenv("MONGO_ENABLE_TLS") == "true",
				Collections: Collections{
					Songs:      os.Getenv("MONGO_COLLECTION_SONGS"),
					Operations: os.Getenv("MONGO_COLLECTION_OPERATIONS"),
				},
			},
		},
	}
}

func LoadConfigAws() *Config {
	region := os.Getenv("AWS_REGION")
	secretName := os.Getenv("AWS_SECRET_NAME")

	sm, err := secretmanager.NewSecretsManager(region)
	if err != nil {
		panic("Error al inicializar el secret manager: " + err.Error())
	}

	secrets, err := sm.GetSecret(context.TODO(), secretName)
	if err != nil {
		panic("Error al obtener los secrets: " + err.Error())
	}
	return &Config{
		Environment: "prod",
		Service: ServiceConfig{
			MaxAttempts: getSecretAsInt(secrets, "SERVICE_MAX_ATTEMPTS", 5),
			Timeout:     time.Duration(getSecretAsInt(secrets, "SERVICE_TIMEOUT", 1)) * time.Minute,
		},

		AWS: AWSConfig{
			Region: region,
		},
		GinConfig: GinConfig{
			Mode: secrets["GIN_MODE"],
		},
		Messaging: MessagingConfig{
			Type: "sqs",
			SQS: &SQSConfig{
				QueueURL: secrets["SQS_QUEUE_URL"],
			},
		},
		API: APIConfig{
			YouTube: YouTubeConfig{
				ApiKey:  secrets["YOUTUBE_API_KEY"],
				Cookies: secrets["COOKIES_YOUTUBE"],
			},
		},
		Storage: StorageConfig{
			Type: "s3-storage",
			S3Config: &S3Config{
				BucketName: secrets["S3_BUCKET_NAME"],
			},
		},
		Database: DatabaseConfig{
			Type: "Dynamodb",
			DynamoDB: &DynamoDBConfig{
				Tables: Tables{
					Songs:      secrets["DYNAMODB_TABLE_SONGS"],
					Operations: secrets["DYNAMODB_TABLE_OPERATIONS"],
				},
			},
		},
	}
}

func getSecretAsInt(secrets map[string]string, key string, defaultValue int) int {
	if valueStr, ok := secrets[key]; ok {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

func getEnvAsInt(name string, defaultValue int) int {
	valueStr := os.Getenv(name)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
