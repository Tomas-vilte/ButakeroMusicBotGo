package config

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/secretmanager"
	"os"
	"strconv"
	"time"
)

// LoadConfig carga la configuración específica para el entorno (local o AWS)
func LoadConfig(environment string) *Config {
	var config Config

	switch environment {
	case "local":
		config = Config{
			Environment: "local",
			Service: ServiceConfig{
				MaxAttempts: getEnvAsInt("SERVICE_MAX_ATTEMPTS", 3),
				Timeout:     time.Duration(getEnvAsInt("SERVICE_TIMEOUT", 1)) * time.Minute,
			},
			GinConfig: GinConfig{
				Mode: os.Getenv("GIN_MODE"),
			},
			Messaging: MessagingConfig{
				Type: "kafka",
				Kafka: &KafkaConfig{
					Brokers: []string{os.Getenv("KAFKA_BROKERS")},
					Topic:   os.Getenv("KAFKA_TOPIC"),
				},
			},
			API: APIConfig{
				YouTube: YouTubeConfig{
					ApiKey: os.Getenv("YOUTUBE_API_KEY"),
				},
				OAuth2: OAuth2Config{
					Enabled: os.Getenv("OAUTH2"),
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
					User:     os.Getenv("MONGO_USER"),
					Password: os.Getenv("MONGO_PASSWORD"),
					Port:     os.Getenv("MONGO_PORT"),
					Host:     []string{os.Getenv("MONGO_HOST")},
					Database: os.Getenv("MONGO_DATABASE"),
					Collections: Collections{
						Songs:      os.Getenv("MONGO_COLLECTION_SONGS"),
						Operations: os.Getenv("MONGO_COLLECTION_OPERATIONS"),
					},
				},
			},
		}

	case "prod":
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

		config = Config{
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
					ApiKey: secrets["YOUTUBE_API_KEY"],
				},
				OAuth2: OAuth2Config{
					Enabled: secrets["OAUTH2"],
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
	default:
		panic("Entorno desconocido: " + environment)
	}

	return &config
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
