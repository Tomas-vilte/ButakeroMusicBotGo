package cache

import (
	"container/list"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/internal/metrics"
	"go.uber.org/zap"
	"sync"
	"time"
)

type ConfigCachingAudio struct {
	MaxCacheSize    int
	CacheTTL        time.Duration
	CleanupInterval time.Duration
}

// DefaultCacheConfigAudio contiene las configuraciones por defecto para el caché.
var DefaultCacheConfigAudio = ConfigCachingAudio{
	MaxCacheSize:    100,
	CacheTTL:        10 * time.Minute,
	CleanupInterval: 5 * time.Minute,
}

type (
	// EntryAudioCaching representa una entrada en el caché de audio con datos binarios y tiempo de expiración.
	EntryAudioCaching struct {
		data     []byte
		expireAt time.Time
	}

	// AudioCaching define métodos para cachear datos de audio.
	AudioCaching interface {
		Get(url string) ([]byte, bool)
		Set(url string, data []byte)
		Size() int
	}

	// AudioCache estructura y métodos de caché de audio.
	AudioCache struct {
		cache         map[string]*list.Element
		config        ConfigCachingAudio
		mu            sync.RWMutex
		stopCleanupCh chan bool
		accessList    ListInterface           // Usamos list.List directamente
		entryPool     EntryPoolInterfaceAudio // Pool para reutilizar las entradas
		logger        logging.Logger          // Interface logging.Logger
		timer         TimerInterface          // Usamos time.Ticker para el temporizador
		metrics       metrics.CacheMetrics
	}
)

// NewAudioCache crea una nueva instancia de AudioCache.
func NewAudioCache(logger logging.Logger, config ConfigCachingAudio, metrics metrics.CacheMetrics) AudioCaching {
	cache := &AudioCache{
		cache:         make(map[string]*list.Element),
		config:        config,
		stopCleanupCh: make(chan bool),
		accessList:    newList(),
		entryPool:     newEntryPoolAudio(),
		logger:        logger,
		timer:         newTimer(config.CleanupInterval),
		metrics:       metrics,
	}

	go cache.startCleanupRoutine()

	return cache
}

// Get recupera los datos de audio almacenados en caché para la URL dada.
// Devuelve los datos y true si la entrada está en caché y no ha expirado; de lo contrario, devuelve nil y false.
func (c *AudioCache) Get(url string) ([]byte, bool) {
	startTime := time.Now()
	c.mu.RLock()
	element, ok := c.cache[url]
	c.mu.RUnlock()

	c.metrics.IncRequests()
	c.metrics.IncGetOperations()

	if !ok {
		c.metrics.IncMisses()
		c.metrics.IncLatencyGet(time.Since(startTime))
		return nil, false
	}

	entry := element.Value.(*EntryAudioCaching)
	if time.Now().After(entry.expireAt) {
		c.deleteEntry(url)
		c.metrics.IncMisses()
		c.metrics.IncLatencyGet(time.Since(startTime))
		return nil, false
	}

	// Actualizar tiempo de acceso en accessList
	c.mu.Lock()
	c.accessList.MoveToFront(element)
	c.mu.Unlock()

	c.metrics.IncHits()
	c.metrics.IncLatencyGet(time.Since(startTime))
	return entry.data, true
}

// Set almacena los datos de audio en caché para la URL dada.
func (c *AudioCache) Set(url string, data []byte) {
	startTime := time.Now()
	expireAt := time.Now().Add(c.config.CacheTTL)

	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.cache[url]; ok {
		entry := element.Value.(*EntryAudioCaching)
		entry.data = data
		entry.expireAt = expireAt
		c.accessList.MoveToFront(element)
		c.metrics.IncSetOperations()
		c.metrics.SetCacheSize(float64(len(c.cache)))
		c.metrics.IncLatencySet(time.Since(startTime))
		c.logger.Info("Datos actualizados en caché de audio para la URL", zap.String("url", url))
	} else {
		entry := c.entryPool.Get()
		entry.data = data
		entry.expireAt = expireAt
		element := c.accessList.PushFront(entry)
		c.cache[url] = element

		if c.accessList.Len() > c.config.MaxCacheSize {
			c.deleteLRUEntry()
		}
		c.metrics.IncSetOperations()
		c.metrics.SetCacheSize(float64(len(c.cache)))
		c.metrics.IncLatencySet(time.Since(startTime))
		c.logger.Info("Datos almacenados en caché de audio para la URL", zap.String("url", url))
	}
	c.metrics.SetCacheSize(float64(len(c.cache)))
}

// deleteEntry elimina la entrada de caché de audio para la URL dada.
func (c *AudioCache) deleteEntry(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, ok := c.cache[url]
	if !ok {
		return
	}

	c.accessList.Remove(element)
	delete(c.cache, url)
	c.entryPool.Put(element.Value.(*EntryAudioCaching))

	c.metrics.IncEvictions()
	c.metrics.SetCacheSize(float64(len(c.cache)))
	c.logger.Info("Entrada de caché de audio eliminada", zap.String("url", url))
}

// deleteLRUEntry elimina la entrada menos recientemente utilizada del caché de audio (LRU).
func (c *AudioCache) deleteLRUEntry() {
	back := c.accessList.Back()
	if back == nil {
		return
	}

	entry := back.Value.(*EntryAudioCaching)
	c.accessList.Remove(back)
	for url, element := range c.cache {
		if element == back {
			delete(c.cache, url)
			break
		}
	}
	c.entryPool.Put(entry)

	c.metrics.IncEvictions()
	c.metrics.SetCacheSize(float64(len(c.cache)))
	c.logger.Info("Entrada de caché de audio LRU eliminada")
}

// startCleanupRoutine inicia el proceso de limpieza periódica de entradas expiradas del caché de audio.
func (c *AudioCache) startCleanupRoutine() {
	defer close(c.stopCleanupCh)
	defer c.timer.Stop()
	for {
		select {
		case <-c.timer.C():
			c.cleanupExpiredEntries()
		case <-c.stopCleanupCh:
			return
		}
	}
}

// cleanupExpiredEntries limpia las entradas expiradas del caché de audio.
func (c *AudioCache) cleanupExpiredEntries() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for url, element := range c.cache {
		entry := element.Value.(*EntryAudioCaching)
		if now.After(entry.expireAt) {
			c.accessList.Remove(element)
			delete(c.cache, url)
			c.entryPool.Put(entry)
			c.metrics.IncEvictions()
			c.logger.Info("Entrada de caché de audio expirada eliminada", zap.String("url", url))
		}
	}
	c.metrics.SetCacheSize(float64(len(c.cache)))
}

func (c *AudioCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	size := len(c.cache)
	c.metrics.SetCacheSize(float64(size))
	return size
}
