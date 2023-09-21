package database

import "go.mongodb.org/mongo-driver/mongo"

type Database struct {
	Name       string
	Connection *mongo.Database
}

func CreateDatabase(client *mongo.Client, name string) *Database {
	return &Database{
		Name:       name,
		Connection: client.Database(name),
	}
}

func (db *Database) CreateCollection(collectionName string) *mongo.Collection {
	return db.Connection.Collection(collectionName)
}
