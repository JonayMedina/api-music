package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/JonayMedina/api-music/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	client *mongo.Client
	db     *mongo.Database
}

func InitDB(cfg *config.Config) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, fmt.Errorf("error conectando a MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("error verificando conexión MongoDB: %v", err)
	}

	return &MongoClient{
		client: client,
		db:     client.Database(cfg.MongoDB),
	}, nil
}

func (m *MongoClient) Close() error {
	return m.client.Disconnect(context.Background())
}

func (m *MongoClient) Collection(name string) *mongo.Collection {
	return m.db.Collection(name)
}
