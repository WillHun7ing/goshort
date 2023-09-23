package database

import "go.mongodb.org/mongo-driver/mongo"

type MongoDB struct {
	Name       string
	Connection *mongo.Database
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
