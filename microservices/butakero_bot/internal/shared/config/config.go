package config

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared"
	"github.com/spf13/viper"
	"time"
)

type (
	Config struct {
		Storage         StorageConfig
		AWS             AWSConfig
		CommandPrefix   string
		Discord         Discord
		Kafka           Kafka
		MongoDB         MongoConfig
		ExternalService ExternalService
	}

	Discord struct {
		Token string
	}

	Kafka struct {
		Brokers []string
		Topic   string
		TLS     shared.TLSConfig
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
	viper.SetDefault("MONGO_DATABASE", "audio_service_db")
	viper.SetDefault("MONGO_COLLECTION", "Songs")
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
		Kafka: Kafka{
			Brokers: viper.GetStringSlice("KAFKA_BROKERS"),
			Topic:   viper.GetString("KAFKA_TOPIC"),
			TLS: shared.TLSConfig{
				Enabled:  viper.GetBool("KAFKA_TLS_ENABLED"),
				CAFile:   viper.GetString("KAFKA_TLS_CA_FILE"),
				CertFile: viper.GetString("KAFKA_TLS_CERT_FILE"),
				KeyFile:  viper.GetString("KAFKA_TLS_KEY_FILE"),
			},
		},
		Discord: Discord{
			Token: viper.GetString("DISCORD_TOKEN"),
		},
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
