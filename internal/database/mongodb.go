package database

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Name       string
	Connection *mongo.Database
}

func CreateClient(mongoUri string) *mongo.Client {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		log.Fatalf("There is some issue regarding connecting to mongodb: %s\n", err)
	}
	return client
}

func CreateDatabase(client *mongo.Client, name string) *MongoDB {
	return &MongoDB{
		Name:       name,
		Connection: client.Database(name),
	}
}

func (db *MongoDB) CreateCollection(collectionName string) *mongo.Collection {
	return db.Connection.Collection(collectionName)
}
