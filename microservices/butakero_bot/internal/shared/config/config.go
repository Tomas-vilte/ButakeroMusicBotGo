package config

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/secretmanager"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared"
	"github.com/spf13/viper"
	"math"
	"strconv"
	"time"
)

type (
	Config struct {
		Storage         StorageConfig
		AWS             AWSConfig
		CommandPrefix   string
		Discord         Discord
		QueueConfig     QueueConfig
		DatabaseConfig  DatabaseConfig
		ExternalService ExternalService
	}

	DatabaseConfig struct {
		MongoDB  MongoConfig
		DynamoDB DynamoDBConfig
	}

	DynamoDBConfig struct {
		SongsTable string
	}

	QueueConfig struct {
		KafkaConfig KafkaConfig
		SQSConfig   SQSConfig
	}

	SQSConfig struct {
		Queues          *QueuesSQS
		MaxMessages     int32
		WaitTimeSeconds int32
	}

	QueuesSQS struct {
		BotDownloadStatusQueueURL   string
		BotDownloadRequestsQueueURL string
	}

	Discord struct {
		Token string
	}

	KafkaConfig struct {
		Brokers []string
		Topics  *KafkaTopics
		TLS     shared.TLSConfig
	}

	KafkaTopics struct {
		BotDownloadStatus  string
		BotDownloadRequest string
	}

	StorageConfig struct {
		S3Config    S3Config
		LocalConfig LocalConfig
	}
	S3Config struct {
		BucketName string
		Region     string
	}

	LocalConfig struct {
		Directory string
	}

	AWSConfig struct {
		Region string
	}

	ExternalService struct {
		BaseURL string
	}

	Host struct {
		Address string
		Port    int
	}
	MongoConfig struct {
		Hosts         []Host
		ReplicaSet    string
		Username      string
		Password      string
		Database      string
		Collection    string
		AuthSource    string
		Timeout       time.Duration
		TLS           shared.TLSConfig
		DirectConnect bool
		RetryWrites   bool
		Port          int
	}
)

func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetDefault("COMMAND_PREFIX", "test")
	viper.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	viper.SetDefault("KAFKA_TOPIC", "notifications")
	viper.SetDefault("MONGO_HOSTS", []string{"localhost"})
	viper.SetDefault("MONGO_PORT", 27017)
	viper.SetDefault("MONGO_DATABASE", "audio_service")
	viper.SetDefault("MONGO_COLLECTION", "songs")
	viper.SetDefault("MONGO_USERNAME", "root")
	viper.SetDefault("MONGO_PASSWORD", "root")
	viper.SetDefault("MONGO_AUTH_SOURCE", "admin")
	viper.SetDefault("MONGO_TIMEOUT", 10*time.Second)
	viper.SetDefault("MONGO_TLS_ENABLED", false)
	viper.SetDefault("MONGO_TLS_CA_FILE", "")
	viper.SetDefault("MONGO_TLS_CERT_FILE", "")
	viper.SetDefault("MONGO_TLS_KEY_FILE", "")
	viper.SetDefault("MONGO_DIRECT_CONNECT", true)
	viper.SetDefault("MONGO_RETRY_WRITES", true)
	viper.SetDefault("LOCAL_STORAGE_DIRECTORY", "/app/data/audio-files")
	viper.SetDefault("AUDIO_PROCESSOR_URL", "http://localhost:8080")

	var mongoHosts []Host
	for _, host := range viper.GetStringSlice("MONGO_HOSTS") {
		mongoHosts = append(mongoHosts, Host{
			Address: host,
			Port:    27017,
		})
	}

	cfg := &Config{
		CommandPrefix: viper.GetString("COMMAND_PREFIX"),
		QueueConfig: QueueConfig{
			KafkaConfig: KafkaConfig{
				Brokers: viper.GetStringSlice("KAFKA_BROKERS"),
				Topics: &KafkaTopics{
					BotDownloadRequest: viper.GetString("KAFKA_BOT_DOWNLOAD_REQUESTS"),
					BotDownloadStatus:  viper.GetString("KAFKA_BOT_DOWNLOAD_STATUS"),
				},
				TLS: shared.TLSConfig{
					Enabled:  viper.GetBool("KAFKA_TLS_ENABLED"),
					CAFile:   viper.GetString("KAFKA_TLS_CA_FILE"),
					CertFile: viper.GetString("KAFKA_TLS_CERT_FILE"),
					KeyFile:  viper.GetString("KAFKA_TLS_KEY_FILE"),
				},
			},
		},
		Discord: Discord{
			Token: viper.GetString("DISCORD_TOKEN"),
		},
		DatabaseConfig: DatabaseConfig{
			MongoDB: MongoConfig{
				Hosts:      mongoHosts,
				ReplicaSet: viper.GetString("MONGO_REPLICA_SET_NAME"),
				Username:   viper.GetString("MONGO_USERNAME"),
				Port:       viper.GetInt("MONGO_PORT"),
				Password:   viper.GetString("MONGO_PASSWORD"),
				Database:   viper.GetString("MONGO_DATABASE"),
				Collection: viper.GetString("MONGO_COLLECTION"),
				AuthSource: viper.GetString("MONGO_AUTH_SOURCE"),
				Timeout:    viper.GetDuration("MONGO_TIMEOUT"),
				TLS: shared.TLSConfig{
					Enabled:  viper.GetBool("MONGO_TLS_ENABLED"),
					CAFile:   viper.GetString("MONGO_TLS_CA_FILE"),
					CertFile: viper.GetString("MONGO_TLS_CERT_FILE"),
					KeyFile:  viper.GetString("MONGO_TLS_KEY_FILE"),
				},
				DirectConnect: viper.GetBool("MONGO_DIRECT_CONNECT"),
				RetryWrites:   viper.GetBool("MONGO_RETRY_WRITES"),
			},
		},
		Storage: StorageConfig{
			LocalConfig: LocalConfig{
				Directory: viper.GetString("LOCAL_STORAGE_DIRECTORY"),
			},
		},
		ExternalService: ExternalService{
			BaseURL: viper.GetString("AUDIO_PROCESSOR_URL"),
		},
	}

	return cfg, nil
}

func LoadConfigAws() (*Config, error) {
	viper.AutomaticEnv()

	region := viper.GetString("AWS_REGION")
	secretName := viper.GetString("AWS_SECRET_NAME")

	sm, err := secretmanager.NewSecretsManager(region)
	if err != nil {
		return nil, fmt.Errorf("error al inicializar secret manager: %w", err)
	}

	secrets, err := sm.GetSecret(context.TODO(), secretName)
	if err != nil {
		return nil, fmt.Errorf("error al obtener secrets: %w", err)
	}

	cfg := &Config{
		CommandPrefix: secrets["COMMAND_PREFIX"],
		Discord: Discord{
			Token: secrets["DISCORD_TOKEN"],
		},
		Storage: StorageConfig{
			S3Config: S3Config{
				BucketName: secrets["S3_BUCKET_NAME"],
				Region:     region,
			},
		},
		AWS: AWSConfig{
			Region: region,
		},
		DatabaseConfig: DatabaseConfig{
			DynamoDB: DynamoDBConfig{
				SongsTable: secrets["DYNAMODB_TABLE_SONGS"],
			},
		},
		ExternalService: ExternalService{
			BaseURL: secrets["AUDIO_PROCESSOR_URL"],
		},
		QueueConfig: QueueConfig{
			SQSConfig: SQSConfig{
				Queues: &QueuesSQS{
					BotDownloadRequestsQueueURL: secrets["SQS_BOT_DOWNLOAD_REQUESTS_URL"],
					BotDownloadStatusQueueURL:   secrets["SQS_BOT_DOWNLOAD_STATUS_URL"],
				},
				MaxMessages:     getSecretAsInt(secrets, "SQS_MAX_MESSAGES", 10),
				WaitTimeSeconds: getSecretAsInt(secrets, "SQS_WAIT_TIME_SECONDS", 20),
			},
		},
	}
	return cfg, nil
}

func getSecretAsInt(secrets map[string]string, key string, defaultValue int32) int32 {
	if valueStr, ok := secrets[key]; ok {
		if value, err := strconv.ParseInt(valueStr, 10, 32); err == nil {
			if value >= math.MinInt32 && value <= math.MaxInt32 {
				return int32(value)
			}
		}
	}
	return defaultValue
}
