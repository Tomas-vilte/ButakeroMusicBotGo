package api

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"strings"
	"time"
)

func CheckMongoDB(ctx context.Context, cfgApplication *config.Config) error {
	uri := buildMongoURI(cfgApplication)

	clientOptions := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(10 * time.Second).
		SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("erorr conectando a MongoDB: %w", err)
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, readpref.PrimaryPreferred())
	if err != nil {
		return fmt.Errorf("error al hacer ping a MongoDB: %w", err)
	}

	cmd := bson.D{{"replSetGetStatus", 1}}
	var result bson.M
	err = client.Database("admin").RunCommand(ctx, cmd).Decode(&result)
	if err != nil {
		return fmt.Errorf("error al obtener estado del replica set: %w", err)
	}

	members, ok := result["members"].(primitive.A)
	if !ok {
		return fmt.Errorf("formato inesperado en la respuesta del replica set")
	}

	hasPrimary := false
	for _, member := range members {
		m := member.(bson.M)
		if state, ok := m["stateStr"].(string); ok && state == "PRIMARY" {
			hasPrimary = true
			break
		}
	}

	if !hasPrimary {
		return fmt.Errorf("no se encontró un nodo primario en el replica set")
	}

	return nil
}

func buildMongoURI(cfg *config.Config) string {
	hostList := strings.Join(cfg.Database.Mongo.Host, ",")
	return fmt.Sprintf("mongodb://%s:%s@%s/?replicaSet=rs0",
		cfg.Database.Mongo.User,
		cfg.Database.Mongo.Password,
		hostList)
}
