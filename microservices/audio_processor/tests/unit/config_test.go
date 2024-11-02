package unit

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestLoadConfig_ValidConfig(t *testing.T) {
	// Crear un archivo de configuración temporal
	configContent := `environment: "local"
gin:
  mode: release
service:
  max_attempts: 3
  timeout: "4m"
api:
  youtube:
    api_key: "test_api_key"
  oauth2:
    enabled: "false"
aws:
  region: "us-east-1"
  credentials:
    access_key: "test_access_key"
    secret_key: "test_secret_key"
messaging:
  type: "kafka"
  kafka:
    brokers: ["localhost:9092"]
    topic: "audio-process-events"
storage:
  type: "local"
  s3:
    bucket_name: "test-bucket"
  local:
    base_path: "data/audio-files"
database:
  type: "mongodb"
  mongodb:
    hosts:
      - "mongo_host01"
      - "mongo_host02"
      - "mongo_host03"
    port: "27017"
    user: "test_user"
    password: "test_password"
    database: "audio_processor"
    collections:
      songs: "songs"
      operations: "operations"
`

	tempFile, err := os.CreateTemp("", "config.yaml")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(configContent)
	require.NoError(t, err)
	tempFile.Close()

	// Cargar la configuración
	cfg, err := config.LoadConfig(tempFile.Name())
	require.NoError(t, err)

	// Verificar la configuración cargada
	assert.Equal(t, "local", cfg.Environment)
	assert.Equal(t, "release", cfg.GinConfig.Mode)
	assert.Equal(t, 3, cfg.Service.MaxAttempts)
	assert.Equal(t, "4m0s", cfg.Service.Timeout.String())
	assert.Equal(t, "us-east-1", cfg.AWS.Region)
	assert.Equal(t, "test_access_key", cfg.AWS.Credentials.AccessKey)
	assert.Equal(t, "test_secret_key", cfg.AWS.Credentials.SecretKey)
	assert.Equal(t, "kafka", cfg.Messaging.Type)
	assert.Equal(t, "localhost:9092", cfg.Messaging.Kafka.Brokers[0])
	assert.Equal(t, "audio-process-events", cfg.Messaging.Kafka.Topic)
	assert.Equal(t, "local", cfg.Storage.Type)
	assert.Equal(t, "test-bucket", cfg.Storage.S3Config.BucketName)
	assert.Equal(t, "data/audio-files", cfg.Storage.LocalConfig.BasePath)
	assert.Equal(t, "mongodb", cfg.Database.Type)
	assert.Equal(t, "mongo_host01", cfg.Database.Mongo.Host[0])
	assert.Equal(t, "27017", cfg.Database.Mongo.Port)
	assert.Equal(t, "test_user", cfg.Database.Mongo.User)
	assert.Equal(t, "test_password", cfg.Database.Mongo.Password)
	assert.Equal(t, "audio_processor", cfg.Database.Mongo.Database)
	assert.Equal(t, "songs", cfg.Database.Mongo.Collections.Songs)
	assert.Equal(t, "operations", cfg.Database.Mongo.Collections.Operations)
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &config.Config{
		Environment: "local",
		Service: config.ServiceConfig{
			MaxAttempts: 3,
			Timeout:     4 * time.Minute,
		},
		GinConfig: config.GinConfig{
			Mode: "release",
		},
		AWS: &config.AWSConfig{
			Region: "us-east-1",
			Credentials: config.CredentialsConfig{
				AccessKey: "test_access_key",
				SecretKey: "test_secret_key",
			},
		},
		Messaging: config.MessagingConfig{
			Type: "kafka",
			Kafka: &config.KafkaConfig{
				Brokers: []string{"localhost:9092"},
				Topic:   "audio-process-events",
			},
		},
		Storage: config.StorageConfig{
			Type: "local",
			LocalConfig: &config.LocalConfig{
				BasePath: "data/audio-files",
			},
		},
		Database: config.DatabaseConfig{
			Type: "mongodb",
			Mongo: &config.MongoConfig{
				Host:     []string{"localhost"},
				Port:     "27017",
				User:     "test_user",
				Password: "test_password",
				Database: "audio_processor",
				Collections: config.Collections{
					Songs:      "songs",
					Operations: "operations",
				},
			},
		},
		API: config.APIConfig{
			YouTube: config.YouTubeConfig{
				ApiKey: "test_api_key",
			},
		},
	}

	// Validar la configuración
	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_InvalidConfig(t *testing.T) {
	cfg := &config.Config{
		Environment: "local",
		Service: config.ServiceConfig{
			MaxAttempts: 0,  // Invalid
			Timeout:     -1, // Invalid
		},
	}

	// Validar la configuración
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max_attempts tiene que ser mayor a 0")
	assert.Contains(t, err.Error(), "timeout tiene que ser mayor a 0")
}
