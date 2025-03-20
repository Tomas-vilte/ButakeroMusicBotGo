package mongodb

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"strings"
	"time"
)

type (
	// Host representa un servidor MongoDB
	Host struct {
		Address string
		Port    int
	}
	// MongoConfig contiene la configuración necesaria para la conexión
	MongoConfig struct {
		Hosts         []Host
		ReplicaSet    string
		Username      string
		Password      string
		Database      string
		AuthSource    string
		Timeout       time.Duration
		TLS           shared.TLSConfig
		DirectConnect bool
		RetryWrites   bool
	}

	// ConnectionManager maneja la conexión a MongoDB
	ConnectionManager struct {
		config MongoConfig
		client *mongo.Client
		logger logging.Logger
	}
)

// NewConnectionManager crea una nueva instancia del administrador de conexiones
func NewConnectionManager(config MongoConfig, logger logging.Logger) *ConnectionManager {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	return &ConnectionManager{
		config: config,
		logger: logger,
	}
}

// setupTLS configura los certificados TLS
func (cm *ConnectionManager) setupTLS() (*tls.Config, error) {
	if !cm.config.TLS.Enabled {
		cm.logger.Info("TLS no está habilitado")
		return nil, nil
	}

	cm.logger.Info("Configurando TLS para MongoDB")
	tlsConfig, err := shared.ConfigureTLS(shared.TLSConfig{
		Enabled:  cm.config.TLS.Enabled,
		CAFile:   cm.config.TLS.CAFile,
		CertFile: cm.config.TLS.CertFile,
		KeyFile:  cm.config.TLS.KeyFile,
	})
	if err != nil {
		cm.logger.Error("Error al configurar TLS", zap.Error(err))
		return nil, fmt.Errorf("error al configurar TLS: %w", err)
	}

	cm.logger.Info("Configuración TLS completada")
	return tlsConfig, nil
}

// Connect establece la conexión con MongoDB
func (cm *ConnectionManager) Connect(ctx context.Context) error {
	uri := cm.buildConnectionString()
	cm.logger.Info("Intentando conectar a MongoDB...")

	clientOptions := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(cm.config.Timeout).
		SetDirect(cm.config.DirectConnect).
		SetRetryWrites(cm.config.RetryWrites).
		SetServerSelectionTimeout(5 * time.Second)

	if cm.config.TLS.Enabled {
		tlsConfig, err := cm.setupTLS()
		if err != nil {
			cm.logger.Error("Error al configurar TLS", zap.Error(err))
			return fmt.Errorf("error al setear TLS: %w", err)
		}
		clientOptions.SetTLSConfig(tlsConfig)
	}
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		cm.logger.Error("Error al conectar al cliente de MongoDB", zap.Error(err))
		return fmt.Errorf("error al conectar a MongoDB client: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		cm.logger.Error("Error al hacer ping a MongoDB", zap.Error(err))
		return fmt.Errorf("error al hacer ping a MongoDB: %w", err)
	}
	cm.client = client
	cm.logger.Info("Conexión a MongoDB establecida exitosamente")
	return nil
}

// Disconnect cierra la conexión con MongoDB
func (cm *ConnectionManager) Disconnect(ctx context.Context) error {
	if cm.client != nil {
		cm.logger.Info("Desconectando de MongoDB")
		if err := cm.client.Disconnect(ctx); err != nil {
			cm.logger.Error("Error al desconectar de MongoDB", zap.Error(err))
			return fmt.Errorf("error al disconectar desde MongoDB: %w", err)
		}
		cm.logger.Info("Desconexión de MongoDB completada")
	}
	return nil
}

// GetClient retorna el cliente de MongoDB
func (cm *ConnectionManager) GetClient() *mongo.Client {
	return cm.client
}

// GetDatabase retorna una base de datos específica
func (cm *ConnectionManager) GetDatabase() *mongo.Database {
	return cm.client.Database(cm.config.Database)
}

// buildConnectionString construye la cadena de conexión para múltiples hosts
func (cm *ConnectionManager) buildConnectionString() string {
	var hosts []string
	for _, host := range cm.config.Hosts {
		hosts = append(hosts, fmt.Sprintf("%s:%d", host.Address, host.Port))
	}

	uri := fmt.Sprintf("mongodb://%s", strings.Join(hosts, ","))

	if cm.config.Username != "" && cm.config.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s",
			cm.config.Username,
			cm.config.Password,
			strings.Join(hosts, ","))
	}

	var opts []string

	if cm.config.ReplicaSet != "" {
		opts = append(opts, fmt.Sprintf("replicaSet=%s", cm.config.ReplicaSet))
	}

	if cm.config.AuthSource != "" {
		opts = append(opts, fmt.Sprintf("authSource=%s", cm.config.AuthSource))
	}

	if len(opts) > 0 {
		uri = fmt.Sprintf("%s/?%s", uri, strings.Join(opts, "&"))
	}
	return uri
}
