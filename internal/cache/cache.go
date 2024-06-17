package cache

import (
	"container/list"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/internal/metrics"
	"go.uber.org/zap"
	"sync"
	"time"
)

type ConfigCaching struct {
	MaxCacheSize    int
	CacheTTL        time.Duration
	CleanupInterval time.Duration
}

// DefaultCacheConfig contiene las configuraciones para el caché.
var DefaultCacheConfig = ConfigCaching{
	MaxCacheSize:    100,
	CacheTTL:        10 * time.Minute,
	CleanupInterval: 5 * time.Minute,
}

type (
	// Entry representa una entrada en el caché con datos de búsqueda y la marca de tiempo de la última actualización.
	Entry struct {
		Results     []*voice.Song
		LastUpdated time.Time
	}

	// Manager define el comportamiento de un caché para almacenar y recuperar resultados de búsqueda.
	Manager interface {
		Get(key string) []*voice.Song
		Set(key string, results []*voice.Song)
		DeleteExpiredEntries()
		Size() int
	}

	// Cache es una implementación simple de caché en memoria.
	Cache struct {
		mu         sync.RWMutex
		Lookup     map[string]*list.Element
		accessList ListInterface      // Interface ListInterface
		entryPool  EntryPoolInterface // Interface EntryPoolInterface
		logger     logging.Logger     // Interface logging.Logger
		stopChan   chan bool
		metrics    metrics.CacheMetrics // Interface metrics.CacheMetrics
		timer      TimerInterface       // Interface TimerInterface
		config     ConfigCaching        // Configurcacion ConfigCaching
	}
)

// NewCache crea una nueva instancia de Cache.
func NewCache(logger logging.Logger, metricsCache metrics.CacheMetrics, config ConfigCaching) Manager {
	cache := &Cache{
		Lookup:     make(map[string]*list.Element),
		accessList: newList(),
		entryPool:  newEntryPool(),
		logger:     logger,
		stopChan:   make(chan bool),
		metrics:    metricsCache,
		timer:      newTimer(config.CleanupInterval),
		config:     config,
	}
	go cache.cleanupExpiredEntries()
	return cache
}

func (c *Cache) Get(key string) []*voice.Song {
	start := time.Now() // Inicio del temporizador para medir la latencia

	c.mu.RLock()
	defer c.mu.RUnlock()

	c.metrics.IncRequests()
	c.metrics.IncGetOperations()
	if element, ok := c.Lookup[key]; ok {
		entry := element.Value.(*Entry)

		c.accessList.MoveToFront(element)
		if time.Since(entry.LastUpdated) > c.config.CacheTTL {
			c.logger.Info("Datos en caché expirados para la entrada", zap.String("input", key))
			c.accessList.Remove(element)
			delete(c.Lookup, key)
			c.entryPool.Put(entry)
			c.metrics.IncEvictions()
			c.metrics.SetCacheSize(float64(len(c.Lookup)))
			c.metrics.IncLatencyGet(time.Since(start)) // Latencia de GET cuando la entrada está expirada
			return nil
		}
		c.metrics.IncLatencyGet(time.Since(start)) // Latencia de GET cuando la entrada es válida
		return entry.Results
	}
	c.logger.Info("Datos en caché no encontrados para la entrada", zap.String("input", key))
	c.metrics.IncLatencyGet(time.Since(start)) // Latencia de GET cuando la entrada no se encuentra
	return nil
}

func (c *Cache) Set(key string, results []*voice.Song) {
	start := time.Now() // Inicio del temporizador para medir la latencia

	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.Lookup[key]; ok {
		entry := element.Value.(*Entry)
		entry.Results = results
		entry.LastUpdated = time.Now()
		c.accessList.MoveToFront(element)
		c.metrics.IncSetOperations()
		c.metrics.SetCacheSize(float64(len(c.Lookup)))
		c.metrics.IncLatencySet(time.Since(start)) // Latencia de SET cuando la entrada es actualizada
		c.logger.Info("Datos actualizados en caché para la entrada", zap.String("input", key))
		return
	}

	newEntry := c.entryPool.Get()
	newEntry.Results = results
	newEntry.LastUpdated = time.Now()
	element := c.accessList.PushFront(newEntry)
	c.Lookup[key] = element

	if c.accessList.Len() > c.config.MaxCacheSize {
		c.DeleteLRUEntry()
	}
	c.logger.Info("Datos almacenados en caché para la entrada", zap.String("input", key))
	c.metrics.IncSetOperations()
	c.metrics.SetCacheSize(float64(len(c.Lookup)))
	c.metrics.IncLatencySet(time.Since(start)) // Latencia de SET cuando la entrada es nueva
}

func (c *Cache) DeleteLRUEntry() {
	if c.accessList.Len() == 0 {
		return
	}

	element := c.accessList.Back()
	if element == nil {
		return
	}

	deleteKey := ""
	for key, e := range c.Lookup {
		if e == element {
			deleteKey = key
			break
		}
	}
	c.accessList.Remove(element)
	delete(c.Lookup, deleteKey)
	c.entryPool.Put(element.Value.(*Entry))
	c.logger.Info("Entrada de caché LRU eliminada", zap.String("input", deleteKey))
	c.metrics.IncEvictions()
	c.metrics.SetCacheSize(float64(len(c.Lookup)))
}

// DeleteExpiredEntries esta función limpia entradas expiradas
func (c *Cache) DeleteExpiredEntries() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, element := range c.Lookup {
		entry := element.Value.(*Entry)
		if now.Sub(entry.LastUpdated) > c.config.CacheTTL {
			c.accessList.Remove(element)
			delete(c.Lookup, key)
			c.entryPool.Put(entry)
			c.logger.Info("Entrada de caché expirada eliminada", zap.String("input", key))
			c.metrics.IncEvictions()
		}
	}
	c.metrics.SetCacheSize(float64(len(c.Lookup)))
}

// Esta función limpia entradas expiradas usando un timer
func (c *Cache) cleanupExpiredEntries() {
	defer close(c.stopChan)
	c.logger.Info("cleanupExpiredEntries iniciada")

	for {
		select {
		case <-c.timer.C():
			c.DeleteExpiredEntries()
			c.timer.Reset(c.config.CleanupInterval)
			c.logger.Info("cleanupExpiredEntries ejecutada")
		case <-c.stopChan:
			c.logger.Info("cleanupExpiredEntries detenida")
			return
		}
	}
}

// Size devuelve el tamaño actual del caché.
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	size := len(c.Lookup)
	c.metrics.SetCacheSize(float64(size))
	return size
}
