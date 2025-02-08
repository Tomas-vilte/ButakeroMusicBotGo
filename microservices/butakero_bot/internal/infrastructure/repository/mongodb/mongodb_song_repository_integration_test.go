package mongodb_test

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/repository/mongodb"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupMongoContainer(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mongo:6",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections").WithStartupTimeout(20 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, "", err
	}

	port, err := container.MappedPort(ctx, "27017")
	if err != nil {
		return nil, "", err
	}

	return container, host + ":" + port.Port(), nil
}

func createTestConnection(t *testing.T, uri string) (*mongo.Client, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+uri))
	require.NoError(t, err)

	return client, func() {
		if err := client.Disconnect(ctx); err != nil {
			return
		}
	}
}

func TestMongoSongRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Saltando prueba de integración en modo corto")
	}

	ctx := context.Background()
	logger, err := logging.NewZapLogger()
	assert.NoError(t, err)

	// Arrange - Configurar Testcontainer
	container, mongoURI, err := setupMongoContainer(ctx)
	assert.NoError(t, err)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			return
		}
	}()

	client, cleanup := createTestConnection(t, mongoURI)
	defer cleanup()

	// Configurar base de datos de prueba
	db := client.Database("test_db")
	collection := db.Collection("songs")
	repo, err := mongodb.NewMongoDBSongRepository(mongodb.Options{
		Collection: collection,
		Logger:     logger,
	})
	assert.NoError(t, err)

	// Test data
	testSong := &entity.Song{
		ID:         "song1",
		VideoID:    "video123",
		Title:      "Test Song",
		Duration:   "3:45",
		URLYoutube: "https://youtube.com/video123",
	}

	t.Run("Obtener canción existente", func(t *testing.T) {
		// Arrange - Insertar datos de prueba
		_, err := collection.InsertOne(ctx, testSong)
		assert.NoError(t, err)
		defer func() {
			if _, err := collection.DeleteOne(ctx, bson.M{"_id": testSong.ID}); err != nil {
				return
			}
		}()

		// Act
		result, err := repo.GetSongByID(ctx, "song1")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, testSong.ID, result.ID)
		assert.Equal(t, testSong.Title, result.Title)
	})

	t.Run("Obtener canción no existente", func(t *testing.T) {
		// Act
		result, err := repo.GetSongByID(ctx, "nonexistent")

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Manejo de errores de base de datos", func(t *testing.T) {
		// Arrange - Forzar error cerrando la conexión
		cleanup()

		// Act
		_, err := repo.GetSongByID(ctx, "song1")

		// Assert
		assert.Error(t, err)
	})
}

func TestConnectionManagerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Saltando prueba de integración en modo corto")
	}

	ctx := context.Background()
	logger, err := logging.NewZapLogger()
	assert.NoError(t, err)

	// Arrange - Configurar Testcontainer
	container, mongoURI, err := setupMongoContainer(ctx)
	require.NoError(t, err)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			return
		}
	}()

	t.Run("Conexión exitosa", func(t *testing.T) {
		host, portStr, _ := strings.Cut(mongoURI, ":")
		port, _ := strconv.Atoi(portStr)
		// Configuración de conexión
		config := mongodb.MongoConfig{
			Hosts:    []mongodb.Host{{Address: host, Port: port}},
			Database: "test_db",
			Timeout:  5 * time.Second,
		}

		// Crear ConnectionManager
		cm := mongodb.NewConnectionManager(config, logger)

		// Act
		err := cm.Connect(ctx)
		defer func() {
			if err := cm.Disconnect(ctx); err != nil {
				return
			}
		}()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, cm.GetClient())
	})

	t.Run("Conexión fallida con credenciales inválidas", func(t *testing.T) {
		config := mongodb.MongoConfig{
			Hosts:    []mongodb.Host{{Address: "localhost", Port: 27017}},
			Username: "in	valid",
			Password: "wrongpassword",
			Database: "test_db",
			Timeout:  2 * time.Second,
		}

		cm := mongodb.NewConnectionManager(config, logger)

		err := cm.Connect(ctx)
		defer func() {
			if err := cm.Disconnect(ctx); err != nil {
				return
			}
		}()

		assert.Error(t, err)
	})
}
