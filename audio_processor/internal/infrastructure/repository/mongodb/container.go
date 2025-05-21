package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBContainer encapsula la configuración del contenedor de MongoDB
type MongoDBContainer struct {
	container testcontainers.Container
	URI       string
	Database  string
	Client    *mongo.Client
}

// MongoDBContainerConfig contiene la configuración para crear un contenedor
type MongoDBContainerConfig struct {
	ImageName      string
	Database       string
	Port           string
	StartupTimeout time.Duration
}

// DefaultMongoDBConfig retorna una configuración por defecto
func DefaultMongoDBConfig() MongoDBContainerConfig {
	return MongoDBContainerConfig{
		ImageName:      "mongo:6",
		Database:       "test_db",
		Port:           "27017/tcp",
		StartupTimeout: 30 * time.Second,
	}
}

// NewMongoDBContainer crea una nueva instancia del contenedor
func NewMongoDBContainer(ctx context.Context, config MongoDBContainerConfig) (*MongoDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        config.ImageName,
		ExposedPorts: []string{config.Port},
		WaitingFor:   wait.ForLog("Waiting for connections").WithStartupTimeout(config.StartupTimeout),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating container: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "27017")
	if err != nil {
		return nil, fmt.Errorf("error getting mapped port: %w", err)
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting host: %w", err)
	}

	uri := fmt.Sprintf("mongodb://%s:%s", hostIP, mappedPort.Port())

	return &MongoDBContainer{
		container: container,
		URI:       uri,
		Database:  config.Database,
	}, nil
}

// Connect establece la conexión con MongoDB
func (m *MongoDBContainer) Connect(ctx context.Context) error {
	clientOptions := options.Client().ApplyURI(m.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("error connecting to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("error pinging MongoDB: %w", err)
	}

	m.Client = client
	return nil
}

// GetCollection retorna una colección específica
func (m *MongoDBContainer) GetCollection(name string) *mongo.Collection {
	return m.Client.Database(m.Database).Collection(name)
}

// Cleanup limpia los recursos del contenedor
func (m *MongoDBContainer) Cleanup(ctx context.Context) error {
	if m.Client != nil {
		if err := m.Client.Disconnect(ctx); err != nil {
			return fmt.Errorf("error disconnecting from MongoDB: %w", err)
		}
	}
	return m.container.Terminate(ctx)
}
