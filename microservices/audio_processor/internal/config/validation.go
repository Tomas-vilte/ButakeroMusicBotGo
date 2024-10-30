package config

import (
	"fmt"
	"strings"
)

func (c *Config) Validate() error {
	if err := c.validateEnvironment(); err != nil {
		return fmt.Errorf("environment error: %w", err)
	}

	if err := c.Service.Validate(); err != nil {
		return fmt.Errorf("service config error: %w", err)
	}

	if c.Environment == "aws" {
		if err := c.validateAWSConfig(); err != nil {
			return fmt.Errorf("aws config error: %w", err)
		}
	}

	if err := c.Messaging.Validate(); err != nil {
		return fmt.Errorf("messaging config error: %w", err)
	}

	if err := c.Storage.Validate(); err != nil {
		return fmt.Errorf("storage config error: %w", err)
	}

	if err := c.Database.Validate(c.Environment); err != nil {
		return fmt.Errorf("database config error: %w", err)
	}

	if err := c.API.Validate(); err != nil {
		return fmt.Errorf("api config error: %w", err)
	}

	if err := c.GinConfig.Validate(); err != nil {
		return fmt.Errorf("gin config validation error: %w", err)
	}

	return nil
}

func (c *Config) validateEnvironment() error {
	switch c.Environment {
	case "local", "aws":
		return nil
	default:
		return fmt.Errorf("environment invalido: %s (puede ser 'local' or 'aws')", c.Environment)
	}
}

func (sc *ServiceConfig) Validate() error {
	var errors []string

	if sc.MaxAttempts <= 0 {
		errors = append(errors, "max_attempts tiene que ser mayor a 0")
	}

	if sc.Timeout <= 0 {
		errors = append(errors, "timeout tiene que ser mayor a 0")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

func (c *Config) validateAWSConfig() error {
	if c.AWS == nil {
		return fmt.Errorf("configuracion de aws es necesaria para entorno de aws")
	}

	var errors []string

	if c.AWS.Region == "" {
		errors = append(errors, "region es necesario")
	}

	if c.AWS.Credentials.AccessKey == "" {
		errors = append(errors, "access_key es necesario")
	}

	if c.AWS.Credentials.SecretKey == "" {
		errors = append(errors, "secret_key es necesario")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

func (mc *MessagingConfig) Validate() error {
	switch mc.Type {
	case "kafka":
		if mc.Kafka == nil {
			return fmt.Errorf("configuracion de kafka es necesaria para el tipo kafka")
		}
		return mc.Kafka.Validate()
	case "sqs":
		if mc.SQS == nil {
			return fmt.Errorf("configuracion de sqs necesaria para el tipo sqs")
		}
		return mc.SQS.Validate()
	default:
		return fmt.Errorf("tipo de mensajeria invalido: %s", mc.Type)
	}
}

func (kc *KafkaConfig) Validate() error {
	var errors []string

	if len(kc.Brokers) == 0 {
		errors = append(errors, "Se requiere al menos un broker de kafka")
	}

	if kc.Topic == "" {
		errors = append(errors, "topic es necesario")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

func (sc *SQSConfig) Validate() error {
	var errors []string

	if sc.QueueURL == "" {
		errors = append(errors, "queue_url es necesario")
	}

	if sc.Topic == "" {
		errors = append(errors, "topic es necesario")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

func (sc *StorageConfig) Validate() error {
	switch sc.Type {
	case "s3":
		if sc.S3Config == nil {
			return fmt.Errorf("configuracion de s3 necesaria para el tipo s3")
		}
		return sc.S3Config.Validate()
	case "local":
		return nil
	default:
		return fmt.Errorf("tipo de storage invalido: %s", sc.Type)
	}
}

func (s3c *S3Config) Validate() error {
	if s3c.BucketName == "" {
		return fmt.Errorf("bucket_name es necesario")
	}
	return nil
}

func (dc *DatabaseConfig) Validate(env string) error {
	switch dc.Type {
	case "mongodb":
		if dc.Mongo == nil {
			return fmt.Errorf("configuracion de mongodb es necesario para el tipo mongodb")
		}
		return dc.Mongo.Validate()
	case "dynamodb":
		if dc.DynamoDB == nil {
			return fmt.Errorf("configuracion de dynamodb es necesario para el tipo dynamodb")
		}
		return dc.DynamoDB.Validate()
	default:
		return fmt.Errorf("tipo de base de datos invalido: %s", dc.Type)
	}
}

func (mc *MongoConfig) Validate() error {
	var errors []string

	if mc.Database == "" {
		errors = append(errors, "database es necesario")
	}

	if mc.Collections.Songs == "" {
		errors = append(errors, "songs es necesario")
	}

	if mc.Collections.Operations == "" {
		errors = append(errors, "operations es necesario")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

func (dc *DynamoDBConfig) Validate() error {
	var errors []string

	if dc.Tables.Songs == "" {
		errors = append(errors, "nombre de la tabla songs es necesario")
	}

	if dc.Tables.Operations == "" {
		errors = append(errors, "nombre de la tabla operations es necesario")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

func (ac *APIConfig) Validate() error {
	var errors []string

	if ac.YouTube.ApiKey == "" {
		errors = append(errors, "YouTube API Key es necesaria")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}

func (g *GinConfig) Validate() error {
	var errors []string

	if g.Mode == "" {
		errors = append(errors, "mode es necesario")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}
	return nil
}
