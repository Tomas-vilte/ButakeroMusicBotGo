package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/utils"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

func CheckMongoDB(ctx context.Context, cfgApplication *config.Config) (*MongoMetadata, error) {
	uri := utils.BuildMongoURI(cfgApplication)

	var tlsConfig *tls.Config
	var err error

	if cfgApplication.Database.Mongo.EnableTLS {
		tlsConfig, err = utils.NewTLSConfig(&utils.TLSConfig{
			CaFile:   cfgApplication.Database.Mongo.CaFile,
			CertFile: cfgApplication.Database.Mongo.CertFile,
			KeyFile:  cfgApplication.Database.Mongo.KeyFile,
		})
		if err != nil {
			return nil, errors.Wrap(err, "Error configurando conexion de TLS de MongoDB")
		}
	}

	clientOptions := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(30 * time.Second).
		SetConnectTimeout(30 * time.Second)

	if cfgApplication.Database.Mongo.EnableTLS {
		clientOptions.SetTLSConfig(tlsConfig)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("error conectando a MongoDB: %w", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			fmt.Printf("Error al desconectar de MongoDB: %v", err)
		}
	}()

	start := time.Now()
	err = client.Ping(ctx, readpref.PrimaryPreferred())
	latencyMs := float64(time.Since(start).Milliseconds())
	if err != nil {
		return nil, fmt.Errorf("error al hacer ping a MongoDB: %w", err)
	}

	var replicaSetStatus bson.M
	if err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "replSetGetStatus", Value: 1}}).Decode(&replicaSetStatus); err != nil {
		return nil, fmt.Errorf("error obteniendo status de replicaSet: %w", err)
	}

	status := ReplicaSetStatus{}
	if members, ok := replicaSetStatus["members"].(primitive.A); ok {
		status.Members = int32(len(members))
		for _, member := range members {
			if m, ok := member.(bson.M); ok {
				if self, ok := m["self"].(bool); ok && self {
					status.Role = m["stateStr"].(string)
					status.Health = m["health"].(float64)
					if electionDate, ok := m["electionDate"].(primitive.DateTime); ok {
						status.LastElection = time.Unix(int64(electionDate)/1000, 0).UTC().Format(time.RFC3339)
					}
					status.SyncStatus = m["syncSourceHost"].(string)
				}
			}
		}
	}
	if setName, ok := replicaSetStatus["set"].(string); ok {
		status.ReplicaSetID = setName
	}

	var buildInfo bson.M
	err = client.Database(cfgApplication.Database.Mongo.Database).RunCommand(ctx, bson.D{{Key: "buildInfo", Value: 1}}).Decode(&buildInfo)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo version de MongoDB: %w", err)
	}
	version, ok := buildInfo["version"].(string)
	if !ok {
		return nil, fmt.Errorf("error obteniendo version de MongoDB")
	}

	var serverStatus bson.M
	err = client.Database(cfgApplication.Database.Mongo.Database).RunCommand(ctx, bson.D{{Key: "serverStatus", Value: 1}}).Decode(&serverStatus)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo status de conexiones de MongoDB: %w", err)
	}
	connections := ConnectionStatus{}
	if conn, ok := serverStatus["connections"].(bson.M); ok {
		connections.Active = getInt32Value(conn, "active")
		connections.Available = getInt32Value(conn, "available")
		connections.Current = getInt32Value(conn, "current")
		connections.Rejected = getInt32Value(conn, "rejected")
	}

	performance := PerformanceStats{
		LatencyMs: latencyMs,
	}
	if opCounters, ok := serverStatus["opcounters"].(bson.M); ok {
		total := int64(0)
		for _, v := range opCounters {
			if count, ok := v.(int64); ok {
				total += count
			}
		}
		performance.OpsPerSec = total
	}

	if mem, ok := serverStatus["mem"].(bson.M); ok {
		if resident, ok := mem["resident"].(int32); ok {
			performance.MemoryUsageMB = resident
		}
	}

	return &MongoMetadata{
		ReplicaSetStatus: status,
		Version:          version,
		Connections:      connections,
		Performance:      performance,
	}, nil
}

func getInt32Value(m bson.M, key string) int32 {
	if value, ok := m[key].(int32); ok {
		return value
	}
	return 0
}
