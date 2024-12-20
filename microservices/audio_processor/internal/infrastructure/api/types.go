package api

// HealthCheckMetadata contiene metadatos para varios servicios utilizados en verificaciones de salud.
type (
	HealthCheckMetadata struct {
		Kafka    *KafkaMetadata    `json:"kafka,omitempty"`    // Metadatos para el servicio Kafka.
		Mongo    *MongoMetadata    `json:"mongo,omitempty"`    // Metadatos para el servicio MongoDB.
		S3       *S3Metadata       `json:"s3,omitempty"`       // Metadatos para el servicio S3.
		DynamoDB *DynamoDBMetadata `json:"dynamodb,omitempty"` // Metadatos para el servicio DynamoDB.
	}

	// DynamoDBMetadata contiene metadatos específicos para DynamoDB.
	DynamoDBMetadata struct {
		TableName string `json:"table_name"` // Nombre de la tabla de DynamoDB.
	}

	// S3Metadata contiene metadatos específicos para S3.
	S3Metadata struct {
		BucketName string `json:"bucket_name"` // Nombre del bucket de S3.
	}

	// KafkaMetadata contiene metadatos específicos para Kafka.
	KafkaMetadata struct {
		Brokers []BrokersMetadata `json:"brokers"` // Lista de brokers de Kafka.
	}

	// BrokersMetadata contiene metadatos para un broker de Kafka.
	BrokersMetadata struct {
		Address  string `json:"address"`   // Dirección del broker de Kafka.
		IsLeader bool   `json:"is_leader"` // Indica si el broker es un líder.
	}

	// MongoMetadata contiene metadatos específicos para MongoDB.
	MongoMetadata struct {
		ReplicaSetStatus ReplicaSetStatus `json:"replica_set_status"` // Estado del conjunto de réplicas de MongoDB.
		Version          string           `json:"version"`            // Versión de MongoDB.
		Connections      ConnectionStatus `json:"connections"`        // Estado de las conexiones de MongoDB.
		Performance      PerformanceStats `json:"performance"`        // Estadísticas de rendimiento de MongoDB.
		Storage          StorageStats     `json:"storage"`            // Estadísticas de almacenamiento de MongoDB.
	}

	// ReplicaSetStatus contiene información de estado para un conjunto de réplicas de MongoDB.
	ReplicaSetStatus struct {
		Role         string  `json:"role"`           // Rol del miembro del conjunto de réplicas.
		Health       float64 `json:"health"`         // Estado de salud del conjunto de réplicas.
		Members      int32   `json:"members"`        // Número de miembros en el conjunto de réplicas.
		LastElection string  `json:"last_election"`  // Marca de tiempo de la última elección.
		ReplicaSetID string  `json:"replica_set_id"` // ID del conjunto de réplicas.
		SyncStatus   string  `json:"sync_status"`    // Estado de sincronización del conjunto de réplicas.
	}

	// ConnectionStatus contiene estadísticas de conexión para MongoDB.
	ConnectionStatus struct {
		Active    int32 `json:"active"`    // Número de conexiones activas.
		Available int32 `json:"available"` // Número de conexiones disponibles.
		Current   int32 `json:"current"`   // Número de conexiones actuales.
		Rejected  int32 `json:"rejected"`  // Número de conexiones rechazadas.
	}

	// PerformanceStats contiene estadísticas de rendimiento para MongoDB.
	PerformanceStats struct {
		LatencyMs     float64 `json:"latency_ms"`      // Latencia en milisegundos.
		OpsPerSec     int64   `json:"ops_per_sec"`     // Operaciones por segundo.
		MemoryUsageMB int32   `json:"memory_usage_mb"` // Uso de memoria en megabytes.
	}

	// StorageStats contiene estadísticas de almacenamiento para MongoDB.
	StorageStats struct {
		DatabaseSizeMB float64 `json:"database_size_mb"` // Tamaño de la base de datos en megabytes.
		DataSizeMB     float64 `json:"data_size_mb"`     // Tamaño de los datos en megabytes.
		IndexSizeMB    float64 `json:"index_size_mb"`    // Tamaño de los índices en megabytes.
	}
)
