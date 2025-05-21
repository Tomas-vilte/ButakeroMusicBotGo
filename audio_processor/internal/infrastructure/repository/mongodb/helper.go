package mongodb

import (
	"context"
	"testing"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/require"
)

// TestHelper encapsula la funcionalidad común para las pruebas
type TestHelper struct {
	MongoDB *MongoDBContainer
	Logger  logger.Logger
	Context context.Context
}

// NewTestHelper crea un nuevo helper para las pruebas
func NewTestHelper(t *testing.T) *TestHelper {
	t.Helper()

	ctx := context.Background()

	// Crear logger
	log, err := logger.NewProductionLogger()
	require.NoError(t, err, "Error creating logger")

	// Crear contenedor con configuración por defecto
	container, err := NewMongoDBContainer(ctx, DefaultMongoDBConfig())
	require.NoError(t, err, "Error creating MongoDB container")

	// Conectar a MongoDB
	err = container.Connect(ctx)
	require.NoError(t, err, "Error connecting to MongoDB")

	return &TestHelper{
		MongoDB: container,
		Logger:  log,
		Context: ctx,
	}
}

// Cleanup limpia los recursos
func (h *TestHelper) Cleanup(t *testing.T) {
	t.Helper()
	if err := h.MongoDB.Cleanup(h.Context); err != nil {
		t.Errorf("Error cleaning up MongoDB container: %v", err)
	}
}
