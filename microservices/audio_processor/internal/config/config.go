package config

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/secretmanager"
	"github.com/spf13/viper"
	"strconv"
	"time"
)

func LoadConfigLocal() *Config {
	viper.AutomaticEnv()

	viper.SetDefault("SERVICE_MAX_ATTEMPTS", 1)
	viper.SetDefault("SERVICE_TIMEOUT", 1)
	viper.SetDefault("GIN_MODE", "debug")
	viper.SetDefault("KAFKA_ENABLE_TLS", false)
	viper.SetDefault("MONGO_ENABLE_TLS", false)
	viper.SetDefault("LOCAL_STORAGE_PATH", "audio-files/")
	viper.SetDefault("ENVIRONMENT", "local")

	return &Config{
		Environment: "local",
		NumWorkers:  2,
		Service: ServiceConfig{
			MaxAttempts: viper.GetInt("SERVICE_MAX_ATTEMPTS"),
			Timeout:     time.Duration(viper.GetInt("SERVICE_TIMEOUT")) * time.Minute,
		},
		GinConfig: GinConfig{
			Mode: viper.GetString("GIN_MODE"),
		},
		Messaging: MessagingConfig{
			Type: "kafka",
			Kafka: &KafkaConfig{
				Brokers: viper.GetStringSlice("KAFKA_BROKERS"),
				Topics: &KafkaTopics{
					BotDownloadStatus:   viper.GetString("KAFKA_BOT_DOWNLOAD_STATUS"),
					BotDownloadRequests: viper.GetString("KAFKA_BOT_DOWNLOAD_REQUESTS"),
				},
				CaFile:    viper.GetString("KAFKA_CA_FILE"),
				CertFile:  viper.GetString("KAFKA_CERT_FILE"),
				KeyFile:   viper.GetString("KAFKA_KEY_FILE"),
				EnableTLS: viper.GetBool("KAFKA_ENABLE_TLS"),
			},
		},
		API: APIConfig{
			YouTube: YouTubeConfig{
				Cookies: viper.GetString("COOKIES_YOUTUBE"),
				ApiKey:  viper.GetString("YOUTUBE_API_KEY"),
			},
		},
		Storage: StorageConfig{
			Type: "local-storage",
			LocalConfig: &LocalConfig{
				BasePath: viper.GetString("LOCAL_STORAGE_PATH"),
			},
		},
		Database: DatabaseConfig{
			Type: "mongodb",
			Mongo: &MongoConfig{
				User:           viper.GetString("MONGO_USER"),
				Password:       viper.GetString("MONGO_PASSWORD"),
				Port:           viper.GetString("MONGO_PORT"),
				Host:           viper.GetStringSlice("MONGO_HOST"),
				CaFile:         viper.GetString("MONGO_CA_FILE"),
				CertFile:       viper.GetString("MONGO_CERT_FILE"),
				KeyFile:        viper.GetString("MONGO_KEY_FILE"),
				Database:       viper.GetString("MONGO_DATABASE"),
				ReplicaSetName: viper.GetString("MONGO_REPLICA_SET_NAME"),
				EnableTLS:      viper.GetBool("MONGO_ENABLE_TLS"),
				Collections: Collections{
					Songs: viper.GetString("MONGO_COLLECTION_SONGS"),
				},
			},
		},
	}
}

func LoadConfigAws() *Config {
	viper.AutomaticEnv()
	region := viper.GetString("AWS_REGION")
	secretName := viper.GetString("AWS_SECRET_NAME")

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
		NumWorkers:  2,
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
					Songs: secrets["DYNAMODB_TABLE_SONGS"],
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
