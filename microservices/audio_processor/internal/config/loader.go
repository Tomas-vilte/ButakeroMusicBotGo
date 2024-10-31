package config

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type (

	// Provider es una interfaz que define como obtener la configuracion
	Provider interface {
		Get() (*Config, error)
	}

	// FileProvider implementa Provider para archivos YAML
	FileProvider struct {
		filePath string
	}
)

// NewFileProvider crea un nuevo proveedor de configuracion basado en archivo
func NewFileProvider(filePath string) *FileProvider {
	return &FileProvider{
		filePath: filePath,
	}
}

func (fp *FileProvider) Get() (*Config, error) {
	cfg, err := fp.readConfig()
	if err != nil {
		return nil, errors.Wrap(err, "error al leer la configuracion")
	}

	cfg.setDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "error en validar configuracion")
	}
	return cfg, nil
}

func (fp *FileProvider) readConfig() (*Config, error) {
	file, err := os.ReadFile(fp.filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "error al leer archivo de configuracion: %s", fp.filePath)
	}

	expandedContent := os.ExpandEnv(string(file))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expandedContent), &cfg); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling config")
	}
	return &cfg, nil
}

func (c *Config) setDefaults() {
	c.setServiceDefaults()
	c.setMessagingDefaults()
	c.setStorageDefaults()
	c.setDatabaseDefaults()
	c.setGinDefaults()
}

func (c *Config) setGinDefaults() {
	if c.GinConfig.Mode == "" {
		c.GinConfig.Mode = gin.DebugMode
	}
}

func (c *Config) setServiceDefaults() {
	if c.Service.MaxAttempts == 0 {
		c.Service.MaxAttempts = 3
	}
	if c.Service.Timeout == 0 {
		c.Service.Timeout = 4 * time.Minute
	}
}

func (c *Config) setMessagingDefaults() {
	if c.Environment == "local" {
		if c.Messaging.Type == "" {
			c.Messaging.Type = "kafka"
		}
		if c.Messaging.Kafka != nil && len(c.Messaging.Kafka.Brokers) == 0 {
			c.Messaging.Kafka.Brokers = []string{"localhost:9092"}
		}
	}
}

func (c *Config) setStorageDefaults() {
	if c.Environment == "local" {
		if c.Storage.Type == "" {
			c.Storage.Type = "local"
		}
	}

	if c.Storage.LocalConfig.BasePath == "" {
		c.Storage.LocalConfig.BasePath = "data/audio-files"
	}
}

func (c *Config) setDatabaseDefaults() {
	if c.Environment == "local" {
		if c.Database.Type == "" {
			c.Database.Type = "mongodb"
		}
		if c.Database.Mongo != nil {
			if c.Database.Mongo.Host == "" {
				c.Database.Mongo.Host = "localhost"
			}
			if c.Database.Mongo.Port == "" {
				c.Database.Mongo.Port = "27017"
			}
			if c.Database.Mongo.Collections.Songs == "" {
				c.Database.Mongo.Collections.Songs = "songs"
			}
			if c.Database.Mongo.Collections.Operations == "" {
				c.Database.Mongo.Collections.Operations = "operations"
			}
		}
	}
}

func (g *GinConfig) ParseBool() bool {
	oauth2Enabled, err := strconv.ParseBool(g.Mode)
	if err != nil {
		return false
	}
	return oauth2Enabled
}

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "config.yaml"
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "no se pudo obtener la ruta de configuración absoluta")
	}

	provider := NewFileProvider(absPath)
	return provider.Get()
}
