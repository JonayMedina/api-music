package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoClient struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoClient(ctx context.Context, uri, dbName string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Crear opciones del cliente
	clientOptions := options.Client().
		ApplyURI(uri).
		SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)).
		SetTimeout(5 * time.Second)

	// Conectar al cliente
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Verificar conexión
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return &MongoClient{
		client: client,
		db:     client.Database(dbName),
	}, nil
}

func (mc *MongoClient) Close(ctx context.Context) error {
	return mc.client.Disconnect(ctx)
}

func (mc *MongoClient) Database() *mongo.Database {
	return mc.db
}

func (mc *MongoClient) Collection(name string) *mongo.Collection {
	return mc.db.Collection(name)
}
