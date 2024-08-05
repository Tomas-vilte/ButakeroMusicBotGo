package cache

import (
	"container/list"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestCache_Get(t *testing.T) {
	t.Run("entrada no encontrada en caché", func(t *testing.T) {
		// Configurar mocks
		listMock := &MockListInterface{}
		loggerMock := new(logging.MockLogger)
		metricsMock := &MockCacheMetrics{}
		entryPoolMock := &MockEntryPoolInterface{}
		timerMock := &MockTimerInterface{}

		// Crear instancia de Cache con los mocks
		cache := &Cache{
			Lookup:     make(map[string]*list.Element),
			accessList: listMock,
			entryPool:  entryPoolMock,
			logger:     loggerMock,
			metrics:    metricsMock,
			timer:      timerMock,
			config:     DefaultCacheConfig,
		}

		// Configurar expectativas de los mocks
		metricsMock.On("IncRequests", mock.Anything).Return()
		metricsMock.On("IncGetOperations", mock.Anything).Return()
		metricsMock.On("IncLatencyGet", mock.Anything, mock.Anything).Return()
		loggerMock.On("Info", "Datos en caché no encontrados para la entrada", mock.Anything).Return()
		metricsMock.On("IncMisses", mock.Anything).Return()

		key := "test_key"
		results := cache.Get(key)
		assert.Nil(t, results)
		metricsMock.AssertNumberOfCalls(t, "IncRequests", 1)
		metricsMock.AssertNumberOfCalls(t, "IncGetOperations", 1)
		loggerMock.AssertCalled(t, "Info", "Datos en caché no encontrados para la entrada", mock.Anything)
	})

	t.Run("entrada encontrada en caché y no expirada", func(t *testing.T) {
		// Configurar mocks
		listMock := &MockListInterface{}
		loggerMock := new(logging.MockLogger)
		metricsMock := &MockCacheMetrics{}
		entryPoolMock := &MockEntryPoolInterface{}
		timerMock := &MockTimerInterface{}

		// Crear instancia de Cache con los mocks
		cache := &Cache{
			Lookup:     make(map[string]*list.Element),
			accessList: listMock,
			entryPool:  entryPoolMock,
			logger:     loggerMock,
			metrics:    metricsMock,
			timer:      timerMock,
			config:     DefaultCacheConfig,
		}

		// Configurar expectativas de los mocks
		metricsMock.On("IncRequests", mock.Anything).Return()
		metricsMock.On("IncGetOperations", mock.Anything).Return()
		metricsMock.On("IncLatencyGet", mock.Anything, mock.Anything).Return()
		metricsMock.On("IncHits", mock.Anything).Return()

		key := "test_key"
		entry := &Entry{
			Results:     []*voice.Song{{Title: "song1"}, {Title: "song2"}},
			LastUpdated: time.Now(),
		}
		element := &list.Element{Value: entry}
		listMock.On("MoveToFront", element).Return()
		cache.Lookup[key] = element

		results := cache.Get(key)
		assert.Equal(t, entry.Results, results)
		metricsMock.AssertNumberOfCalls(t, "IncRequests", 1)
		metricsMock.AssertNumberOfCalls(t, "IncGetOperations", 1)
		listMock.AssertCalled(t, "MoveToFront", element)
	})

	t.Run("entrada encontrada en caché pero expirada", func(t *testing.T) {
		// Configurar mocks
		listMock := &MockListInterface{}
		loggerMock := new(logging.MockLogger)
		metricsMock := &MockCacheMetrics{}
		entryPoolMock := &MockEntryPoolInterface{}
		timerMock := &MockTimerInterface{}
		cache := &Cache{
			Lookup:     make(map[string]*list.Element),
			accessList: listMock,
			entryPool:  entryPoolMock,
			logger:     loggerMock,
			metrics:    metricsMock,
			timer:      timerMock,
			config:     DefaultCacheConfig,
		}
		metricsMock.On("IncRequests", mock.Anything).Return()
		metricsMock.On("IncGetOperations", mock.Anything).Return()
		metricsMock.On("IncEvictions", mock.Anything).Return()
		metricsMock.On("SetCacheSize", mock.Anything, mock.Anything).Return()
		metricsMock.On("IncLatencyGet", mock.Anything, mock.Anything).Return()
		metricsMock.On("IncMisses", mock.Anything).Return()

		loggerMock.On("Info", "Datos en caché expirados para la entrada", mock.Anything).Return()

		key := "test_key"
		entry := &Entry{
			Results:     []*voice.Song{{Title: "song1"}, {Title: "song2"}},
			LastUpdated: time.Now().Add(-2 * DefaultCacheConfig.CacheTTL),
		}
		element := &list.Element{Value: entry}
		listMock.On("Remove", element).Return()
		entryPoolMock.On("Put", entry).Return()
		listMock.On("MoveToFront", element).Return()
		cache.Lookup[key] = element

		results := cache.Get(key)
		assert.Nil(t, results)
		metricsMock.AssertNumberOfCalls(t, "IncRequests", 1)
		metricsMock.AssertNumberOfCalls(t, "IncGetOperations", 1)
		metricsMock.AssertNumberOfCalls(t, "IncEvictions", 1)
		listMock.AssertCalled(t, "Remove", element)
		entryPoolMock.AssertCalled(t, "Put", entry)
		listMock.AssertCalled(t, "MoveToFront", element)
		loggerMock.AssertCalled(t, "Info", "Datos en caché expirados para la entrada", mock.Anything)
	})
}

// Pruebas para la función Set
func TestCache_Set(t *testing.T) {
	t.Run("actualizar entrada existente en caché", func(t *testing.T) {
		// Configurar mocks
		listMock := &MockListInterface{}
		loggerMock := new(logging.MockLogger)
		metricsMock := &MockCacheMetrics{}
		entryPoolMock := &MockEntryPoolInterface{}
		timerMock := &MockTimerInterface{}

		// Crear instancia de Cache con los mocks
		cache := &Cache{
			Lookup:     make(map[string]*list.Element),
			accessList: listMock,
			entryPool:  entryPoolMock,
			logger:     loggerMock,
			metrics:    metricsMock,
			timer:      timerMock,
			config:     DefaultCacheConfig,
		}

		// Configurar expectativas de los mocks
		metricsMock.On("IncSetOperations", mock.Anything).Return()
		metricsMock.On("SetCacheSize", float64(1)).Return()
		metricsMock.On("IncLatencySet", mock.Anything, mock.Anything).Return()

		entry := &Entry{
			Results:     []*voice.Song{{Title: "old_song"}},
			LastUpdated: time.Now(),
		}
		element := &list.Element{Value: entry}
		cache.Lookup["test_key"] = element

		newResults := []*voice.Song{{Title: "new_song"}}
		listMock.On("MoveToFront", element).Return()
		loggerMock.On("Info", "Datos actualizados en caché para la entrada", mock.Anything).Return()

		cache.Set("test_key", newResults)
		assert.Equal(t, newResults, entry.Results)
		listMock.AssertCalled(t, "MoveToFront", element)
		metricsMock.AssertNumberOfCalls(t, "IncSetOperations", 1)
		metricsMock.AssertNumberOfCalls(t, "SetCacheSize", 1)
		loggerMock.AssertCalled(t, "Info", "Datos actualizados en caché para la entrada", mock.Anything)
	})
	t.Run("añadir nueva entrada en caché", func(t *testing.T) {
		// Configurar mocks
		listMock := &MockListInterface{}
		loggerMock := new(logging.MockLogger)
		metricsMock := &MockCacheMetrics{}
		entryPoolMock := &MockEntryPoolInterface{}
		timerMock := &MockTimerInterface{}

		// Crear instancia de Cache con los mocks
		cache := &Cache{
			Lookup:     make(map[string]*list.Element),
			accessList: listMock,
			entryPool:  entryPoolMock,
			logger:     loggerMock,
			metrics:    metricsMock,
			timer:      timerMock,
			config:     DefaultCacheConfig,
		}

		// Configurar expectativas de los mocks
		metricsMock.On("IncSetOperations", mock.Anything).Return()
		metricsMock.On("SetCacheSize", float64(1)).Return()
		metricsMock.On("IncLatencySet", mock.Anything, mock.Anything).Return()
		entryPoolMock.On("Get").Return(&Entry{})

		newResults := []*voice.Song{{Title: "new_song"}}
		// Utilizamos cualquier predicado para el campo `LastUpdated`
		listMock.On("PushFront", mock.MatchedBy(func(e *Entry) bool {
			return assert.ObjectsAreEqual(newResults, e.Results)
		})).Return(&list.Element{Value: &Entry{
			Results: newResults,
		}})
		listMock.On("Len").Return(1)
		loggerMock.On("Info", "Datos almacenados en caché para la entrada", mock.Anything).Return()

		cache.Set("test_key", newResults)
		assert.Equal(t, newResults, cache.Lookup["test_key"].Value.(*Entry).Results)
		listMock.AssertCalled(t, "PushFront", mock.MatchedBy(func(e *Entry) bool {
			return assert.ObjectsAreEqual(newResults, e.Results)
		}))
		metricsMock.AssertNumberOfCalls(t, "IncSetOperations", 1)
		metricsMock.AssertNumberOfCalls(t, "SetCacheSize", 1)
		loggerMock.AssertCalled(t, "Info", "Datos almacenados en caché para la entrada", mock.Anything)
	})
}

func TestCache_Set_DeleteLRUEntry(t *testing.T) {
	// Configurar mocks
	listMock := &MockListInterface{}
	loggerMock := new(logging.MockLogger)
	metricsMock := &MockCacheMetrics{}
	entryPoolMock := &MockEntryPoolInterface{}

	// Crear instancia de Cache con los mocks
	cache := &Cache{
		Lookup:     make(map[string]*list.Element),
		accessList: listMock,
		entryPool:  entryPoolMock,
		logger:     loggerMock,
		metrics:    metricsMock,
		config:     ConfigCaching{MaxCacheSize: 2}, // Tamaño máximo de la caché para probar
	}

	// Simular dos entradas en la caché
	entry1 := &Entry{
		Results:     []*voice.Song{{Title: "song1"}},
		LastUpdated: time.Now(),
	}
	element1 := &list.Element{Value: entry1}
	cache.Lookup["key1"] = element1

	entry2 := &Entry{
		Results:     []*voice.Song{{Title: "song2"}},
		LastUpdated: time.Now(),
	}
	element2 := &list.Element{Value: entry2}
	cache.Lookup["key2"] = element2

	// Configurar expectativas de los mocks
	entryPoolMock.On("Get").Return(&Entry{})
	listMock.On("PushFront", mock.Anything).Return(&list.Element{})
	listMock.On("Len").Return(3) // Mock para indicar que hay 3 elementos en la lista

	// Expectativas para cuando se llama a DeleteLRUEntry
	listMock.On("Back").Return(element1)
	listMock.On("Remove", element1).Return()
	entryPoolMock.On("Put", entry1).Return()
	metricsMock.On("IncEvictions", mock.Anything).Return()
	metricsMock.On("IncLatencySet", mock.Anything, mock.Anything).Return()
	metricsMock.On("SetCacheSize", mock.Anything).Return()
	loggerMock.On("Info", "Entrada de caché LRU eliminada", mock.Anything).Return()
	loggerMock.On("Info", "Datos almacenados en caché para la entrada", mock.Anything).Return()
	metricsMock.On("IncSetOperations", mock.Anything).Return()

	// Simular exceder el tamaño máximo de la caché al agregar una nueva entrada
	newResults := []*voice.Song{{Title: "song3"}}
	cache.Set("key3", newResults)

	// Verificar que se llamó a DeleteLRUEntry para eliminar la entrada LRU
	assert.Equal(t, 2, len(cache.Lookup)) // Debe haber solo 2 entradas después de eliminar la LRU

	// Asegurarse de que se llamaron las funciones necesarias en los mocks
	entryPoolMock.AssertExpectations(t)
	listMock.AssertExpectations(t)
	metricsMock.AssertExpectations(t)
	loggerMock.AssertExpectations(t)
}

func TestCache_DeleteLRUEntry(t *testing.T) {
	// Configurar mocks
	listMock := &MockListInterface{}
	loggerMock := new(logging.MockLogger)
	metricsMock := &MockCacheMetrics{}
	entryPoolMock := &MockEntryPoolInterface{}

	// Crear instancia de Cache con los mocks
	cache := &Cache{
		Lookup:     make(map[string]*list.Element),
		accessList: listMock,
		entryPool:  entryPoolMock,
		logger:     loggerMock,
		metrics:    metricsMock,
		config:     ConfigCaching{MaxCacheSize: 1},
	}

	// Caso 1: accessList.Len() == 0
	listMock.On("Len").Return(0)

	cache.DeleteLRUEntry() // Llamar a DeleteLRUEntry

	listMock.AssertExpectations(t)
	entryPoolMock.AssertExpectations(t)
	metricsMock.AssertExpectations(t)
	loggerMock.AssertExpectations(t)

	// Caso 2: element == nil
	cache.accessList = listMock

	cache.DeleteLRUEntry()

	listMock.AssertExpectations(t)
	entryPoolMock.AssertExpectations(t)
	metricsMock.AssertExpectations(t)
	loggerMock.AssertExpectations(t)
}

func TestCache_DeleteExpiredEntries(t *testing.T) {
	// Configurar mocks
	listMock := &MockListInterface{}
	loggerMock := new(logging.MockLogger)
	metricsMock := &MockCacheMetrics{}
	entryPoolMock := &MockEntryPoolInterface{}

	// Crear instancia de Cache con los mocks
	cache := &Cache{
		Lookup:     make(map[string]*list.Element),
		accessList: listMock,
		entryPool:  entryPoolMock,
		logger:     loggerMock,
		metrics:    metricsMock,
		config:     DefaultCacheConfig,
	}

	// Crear entradas de prueba en la caché
	now := time.Now()
	entry1 := &Entry{
		Results:     []*voice.Song{{Title: "song1"}},
		LastUpdated: now.Add(-2 * time.Hour), // Hace 2 horas
	}
	element1 := &list.Element{Value: entry1}
	cache.Lookup["key1"] = element1

	entry2 := &Entry{
		Results:     []*voice.Song{{Title: "song2"}},
		LastUpdated: now.Add(-1 * time.Hour), // Hace 1 hora
	}
	element2 := &list.Element{Value: entry2}
	cache.Lookup["key2"] = element2

	// Configurar expectativas de los mocks
	listMock.On("Remove", element1).Return()
	listMock.On("Remove", element2).Return()
	entryPoolMock.On("Put", entry1).Return()
	entryPoolMock.On("Put", entry2).Return()
	metricsMock.On("IncEvictions", mock.Anything).Return().Times(1)
	metricsMock.On("SetCacheSize", mock.Anything).Return()
	loggerMock.On("Info", "Entrada de caché expirada eliminada", mock.Anything).Return()
	loggerMock.On("Info", "Entrada de caché expirada eliminada", mock.Anything).Return()

	// Ejecutar la función DeleteExpiredEntries
	cache.DeleteExpiredEntries()

	// Verificar que las entradas expiradas hayan sido eliminadas de la caché
	assert.Equal(t, 0, len(cache.Lookup)) // La caché debe estar vacía después de eliminar las entradas expiradas

	// Asegurarse de que se llamaron las funciones necesarias en los mocks
	listMock.AssertExpectations(t)
	entryPoolMock.AssertExpectations(t)
	metricsMock.AssertExpectations(t)
	loggerMock.AssertExpectations(t)

	// Verificar que se actualizó el tamaño de la caché en las métricas
	metricsMock.AssertCalled(t, "SetCacheSize", mock.Anything)
}

func TestCache_cleanupExpiredEntries(t *testing.T) {
	// Configurar mocks
	listMock := &MockListInterface{}
	loggerMock := new(logging.MockLogger)
	metricsMock := &MockCacheMetrics{}
	entryPoolMock := &MockEntryPoolInterface{}
	timerMock := &MockTimerInterface{}

	// Crear canal de parada
	stopChan := make(chan bool)

	cacheConfig := ConfigCaching{
		MaxCacheSize:    100,
		CacheTTL:        5 * time.Minute,
		CleanupInterval: 10 * time.Second,
	}

	// Crear instancia de Cache con los mocks y el canal de parada
	cache := &Cache{
		Lookup:     make(map[string]*list.Element),
		accessList: listMock,
		entryPool:  entryPoolMock,
		logger:     loggerMock,
		stopChan:   stopChan,
		metrics:    metricsMock,
		timer:      timerMock,
		config:     cacheConfig,
	}

	// Configurar expectativas de los mocks para la limpieza de entradas expiradas
	timerInterval := cacheConfig.CleanupInterval // Intervalo de tiempo esperado
	loggerMock.On("Info", "cleanupExpiredEntries iniciada", mock.Anything).Return()
	loggerMock.On("Info", "cleanupExpiredEntries ejecutada", mock.Anything).Return()
	loggerMock.On("Info", "cleanupExpiredEntries detenida", mock.Anything).Return().Maybe()
	timerMock.On("C").Return(time.After(timerInterval)).Maybe() // Simular tick del timer después del intervalo esperado
	timerMock.On("Reset", timerInterval).Return(true).Maybe()   // Simular reset del timer
	timerMock.On("Stop").Return(true).Maybe()                   // Simular stop del timer
	metricsMock.On("SetCacheSize", mock.Anything).Maybe()

	// Simular ejecución de la limpieza de entradas expiradas
	go cache.cleanupExpiredEntries()

	// Esperar para asegurarse de que la función tenga tiempo de ejecutarse
	time.Sleep(timerInterval + 100*time.Millisecond)

	// Detener la ejecución simulada
	stopChan <- true

	// Asegurarse de que se llamaron las funciones necesarias en los mocks
	timerMock.AssertExpectations(t)
	listMock.AssertExpectations(t)
	entryPoolMock.AssertExpectations(t)
	metricsMock.AssertExpectations(t)
	loggerMock.AssertExpectations(t)
}

func TestCache_Size(t *testing.T) {
	// Configurar mocks
	listMock := &MockListInterface{}
	loggerMock := new(logging.MockLogger)
	metricsMock := &MockCacheMetrics{}
	entryPoolMock := &MockEntryPoolInterface{}
	timerMock := &MockTimerInterface{}

	// Crear instancia de Cache con los mocks
	cache := &Cache{
		Lookup:     make(map[string]*list.Element),
		accessList: listMock,
		entryPool:  entryPoolMock,
		logger:     loggerMock,
		metrics:    metricsMock,
		timer:      timerMock,
		config:     DefaultCacheConfig,
	}

	metricsMock.On("SetCacheSize", mock.Anything).Return()

	t.Run("Cache vacio", func(t *testing.T) {
		size := cache.Size()
		assert.Equal(t, 0, size, "El tamaño del cache debe ser 0 cuando está vacío")
	})

	t.Run("Cache con una entrada", func(t *testing.T) {
		entry := &Entry{
			Results:     []*voice.Song{{Title: "song1"}},
			LastUpdated: time.Now(),
		}
		element := &list.Element{Value: entry}
		cache.Lookup["key1"] = element

		size := cache.Size()
		assert.Equal(t, 1, size, "El tamaño del cache debe ser 1 después de agregar una entrada")
	})

	t.Run("Cache con varias entradas", func(t *testing.T) {
		entry2 := &Entry{
			Results:     []*voice.Song{{Title: "song2"}},
			LastUpdated: time.Now(),
		}
		element2 := &list.Element{Value: entry2}
		cache.Lookup["key2"] = element2

		entry3 := &Entry{
			Results:     []*voice.Song{{Title: "song3"}},
			LastUpdated: time.Now(),
		}
		element3 := &list.Element{Value: entry3}
		cache.Lookup["key3"] = element3

		size := cache.Size()
		assert.Equal(t, 3, size, "El tamaño del cache debe ser 3 después de agregar tres entradas")
	})
}
