package cache

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestCache_Set(t *testing.T) {
	mockLogger := new(MockLogger)
	mockMetrics := new(MockCacheMetrics)
	cache := NewCache(mockLogger, mockMetrics).(*Cache)

	// Configuración de la entrada de caché
	key := "test-key"
	songs := []*voice.Song{
		{Title: "Test Song"},
	}

	// Expectativas
	mockLogger.On("Info", "Datos almacenados en caché para la entrada", mock.Anything).Once()
	mockMetrics.On("IncSetOperations").Once()
	mockMetrics.On("SetCacheSize", mock.Anything).Once()

	// Llamada al método Set
	cache.Set(key, songs)

	// Verificaciones
	mockLogger.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
}

func TestCache_DeleteExpiredEntries(t *testing.T) {
	mockLogger := new(MockLogger)
	mockMetrics := new(MockCacheMetrics)
	cache := NewCache(mockLogger, mockMetrics).(*Cache)

	// Configuración de la entrada de caché
	key := "test-key"
	songs := []*voice.Song{
		{Title: "Test Song"},
	}

	// Expectativas para Set
	mockLogger.On("Info", "Datos almacenados en caché para la entrada", mock.Anything).Once()
	mockMetrics.On("IncSetOperations").Once()
	mockMetrics.On("SetCacheSize", mock.Anything).Once()

	// Llamada al método Set
	cache.Set(key, songs)

	// Modificar LastUpdated para simular entrada expirada
	cache.mu.Lock()
	cache.Lookup[key] = CacheEntry{
		Results:     songs,
		LastUpdated: time.Now().Add(-cacheTTL - 1*time.Minute),
	}
	cache.mu.Unlock()

	// Expectativas para DeleteExpiredEntries
	mockLogger.On("Info", "Entrada de caché expirada eliminada", mock.Anything).Once()
	mockMetrics.On("IncEvictions").Once()
	mockMetrics.On("SetCacheSize", mock.Anything).Once()

	// Llamada al método DeleteExpiredEntries
	cache.DeleteExpiredEntries()

	// Verificaciones
	mockLogger.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
}

func TestCache_Get_Hit(t *testing.T) {
	mockLogger := new(MockLogger)
	mockMetrics := new(MockCacheMetrics)
	cache := NewCache(mockLogger, mockMetrics).(*Cache)

	// Configuración de la entrada de caché
	key := "test-key"
	songs := []*voice.Song{
		{Title: "Test Song"},
	}

	// Expectativas para Set
	mockLogger.On("Info", "Datos almacenados en caché para la entrada", mock.Anything).Once()
	mockMetrics.On("IncSetOperations").Once()
	mockMetrics.On("SetCacheSize", mock.Anything).Once()

	// Llamada al método Set
	cache.Set(key, songs)

	// Expectativas para Get
	mockLogger.AssertNotCalled(t, "Info", mock.Anything)
	mockMetrics.On("IncRequests").Once()
	mockMetrics.On("IncGetOperations").Once()

	// Llamada al método Get
	results := cache.Get(key)

	// Verificaciones
	mockLogger.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
	if len(results) != 1 || results[0].Title != "Test Song" {
		t.Errorf("Resultados incorrectos de Get: %v", results)
	}
}

func TestCache_Get_Miss(t *testing.T) {
	mockLogger := new(MockLogger)
	mockMetrics := new(MockCacheMetrics)
	cache := NewCache(mockLogger, mockMetrics).(*Cache)

	// Configuración de la entrada de caché
	key := "test-key"

	// Expectativas
	mockLogger.On("Info", "Datos en caché expirados o no encontrados para la entrada", mock.Anything).Once()
	mockMetrics.On("IncRequests").Once()
	mockMetrics.On("IncGetOperations").Once()

	// Llamada al método Get
	results := cache.Get(key)

	// Verificaciones
	mockLogger.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
	if results != nil {
		t.Errorf("Resultados incorrectos de Get: %v", results)
	}
}

func TestCache_Get_Expired(t *testing.T) {
	mockLogger := new(MockLogger)
	mockMetrics := new(MockCacheMetrics)
	cache := NewCache(mockLogger, mockMetrics).(*Cache)

	// Configuración de la entrada de caché
	key := "test-key"
	songs := []*voice.Song{
		{Title: "Test Song"},
	}

	// Expectativas para Set
	mockLogger.On("Info", "Datos almacenados en caché para la entrada", mock.Anything).Once()
	mockMetrics.On("IncSetOperations").Once()
	mockMetrics.On("SetCacheSize", mock.Anything).Once()

	cache.Set(key, songs)

	// Modificar LastUpdated para simular entrada expirada
	cache.mu.Lock()
	cache.Lookup[key] = CacheEntry{
		Results:     songs,
		LastUpdated: time.Now().Add(-cacheTTL - 1*time.Minute),
	}
	cache.mu.Unlock()

	// Expectativas
	mockLogger.On("Info", "Datos en caché expirados o no encontrados para la entrada", mock.Anything).Once()
	mockMetrics.On("IncRequests").Once()
	mockMetrics.On("IncGetOperations").Once()

	// Llamada al método Get
	results := cache.Get(key)

	// Verificaciones
	mockLogger.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
	if results != nil {
		t.Errorf("Resultados incorrectos de Get: %v", results)
	}
}

func TestCache_Size(t *testing.T) {
	mockLogger := new(MockLogger)
	mockMetrics := new(MockCacheMetrics)
	cache := NewCache(mockLogger, mockMetrics).(*Cache)

	// Configuración de la entrada de caché
	key := "test-key"
	songs := []*voice.Song{
		{Title: "Test Song"},
	}
	// Expectativas para Set
	mockLogger.On("Info", "Datos almacenados en caché para la entrada", mock.Anything).Once()
	mockMetrics.On("IncSetOperations").Once()
	mockMetrics.On("SetCacheSize", mock.Anything).Once()

	cache.Set(key, songs)

	// Llamada al método Size y verificación de resultados
	size := cache.Size()
	assert.Equal(t, 1, size)
}
