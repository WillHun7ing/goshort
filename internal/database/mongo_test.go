package database

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	testDBName                = "testdb"
	testCollectionName        = "testcollection"
	mongoTestConnectionString = "mongodb://localhost:27017"
)

func TestCreateDatabase(t *testing.T) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoTestConnectionString))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	db := CreateDatabase(client, testDBName)
	if db.Name != testDBName {
		t.Fatalf("Expected db name to be %s, but got %s", testDBName, db.Name)
	}
}

func TestCreateCollection(t *testing.T) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoTestConnectionString))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	db := CreateDatabase(client, testDBName)
	collection := db.CreateCollection(testCollectionName)
	if collection.Name() != testCollectionName {
		t.Fatalf("Expected collection name to be %s, but got %s", testCollectionName, collection.Name())
	}
}
