package cache

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"go.uber.org/zap"
	"sync"
	"time"
)

const cacheTTL = 3 * time.Minute

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
	}

	// Cache es una implementación simple de caché en memoria.
	Cache struct {
		mu     sync.RWMutex
		lookup map[string]CacheEntry
		logger logging.Logger
	}
)

// NewCache crea una nueva instancia de Cache.
func NewCache(logger logging.Logger) CacheManager {
	return &Cache{
		lookup: make(map[string]CacheEntry),
		logger: logger,
	}
}

func (c *Cache) Get(key string) []*voice.Song {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.lookup[key]
	if !ok || time.Since(entry.LastUpdated) > cacheTTL {
		c.logger.Info("Datos en caché expirados o no encontrados para la entrada", zap.String("input", key))
		return nil
	}
	return entry.Results
}

func (c *Cache) Set(key string, results []*voice.Song) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lookup[key] = CacheEntry{
		Results:     results,
		LastUpdated: time.Now(),
	}
	c.logger.Info("Datos almacenados en caché para la entrada", zap.String("input", key))
}
