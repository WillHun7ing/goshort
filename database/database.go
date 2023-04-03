package database

import "go.mongodb.org/mongo-driver/mongo"

type Database struct {
	Name       string
	Connection *mongo.Database
}

func CreateDatabase(client *mongo.Client, name string) *Database {
	connection := client.Database(name)
	db := Database{
		Name:       name,
		Connection: connection,
	}
	return &db
}

func (db *Database) CreateCollection(collectionName string) *mongo.Collection {
	collection := db.Connection.Collection(collectionName)
	return collection
}
