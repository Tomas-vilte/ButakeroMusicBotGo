package cache

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/internal/metrics"
	"go.uber.org/zap"
	"sync"
	"time"
)

const cacheTTL = 5 * time.Minute
const cleanupInterval = 1 * time.Minute

type (
	// CacheEntry representa una entrada en el caché con datos de búsqueda y la marca de tiempo de la última actualización.
	CacheEntry struct {
		Results     []*voice.Song
		LastUpdated time.Time
	}

	// CacheManager define el comportamiento de un caché para almacenar y recuperar resultados de búsqueda.
	CacheManager interface {
		Get(key string) []*voice.Song
		Set(key string, results []*voice.Song)
		DeleteExpiredEntries()
		Size() int
	}

	// Cache es una implementación simple de caché en memoria.
	Cache struct {
		mu       sync.RWMutex
		Lookup   map[string]CacheEntry
		logger   logging.Logger
		stopChan chan bool
		metrics  metrics.CacheMetrics
	}
)

// NewCache crea una nueva instancia de Cache.
func NewCache(logger logging.Logger, metricsCache metrics.CacheMetrics) CacheManager {
	cache := &Cache{
		Lookup:   make(map[string]CacheEntry),
		logger:   logger,
		stopChan: make(chan bool),
		metrics:  metricsCache,
	}
	go cache.cleanupExpiredEntries()
	return cache
}

func (c *Cache) Get(key string) []*voice.Song {
	c.mu.RLock()
	defer c.mu.RUnlock()

	c.metrics.IncRequests()
	c.metrics.IncGetOperations()
	entry, ok := c.Lookup[key]
	if !ok || time.Since(entry.LastUpdated) > cacheTTL {
		c.logger.Info("Datos en caché expirados o no encontrados para la entrada", zap.String("input", key))
		return nil
	}
	c.logger.Info("Dato recuperado desde la cache: ", zap.String("Cancion", key))
	return entry.Results
}

func (c *Cache) Set(key string, results []*voice.Song) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Lookup[key] = CacheEntry{
		Results:     results,
		LastUpdated: time.Now(),
	}
	c.logger.Info("Datos almacenados en caché para la entrada", zap.String("input", key))
	c.metrics.IncSetOperations()
	c.metrics.SetCacheSize(float64(len(c.Lookup)))
}

func (c *Cache) DeleteExpiredEntries() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.Lookup {
		if now.Sub(entry.LastUpdated) > cacheTTL {
			delete(c.Lookup, key)
			c.logger.Info("Entrada de caché expirada eliminada", zap.String("input", key))
			c.metrics.IncEvictions()
		}
	}
	c.metrics.SetCacheSize(float64(len(c.Lookup)))
}

func (c *Cache) cleanupExpiredEntries() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	defer close(c.stopChan)
	for {
		select {
		case <-ticker.C:
			c.DeleteExpiredEntries()
		case <-c.stopChan:
			return
		}
	}
}

// Size devuelve el tamaño actual del caché.
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.Lookup)
}
