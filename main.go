package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	Name    string `bson:"name"`
	Surname string `bson:"surname"`
}

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	err = mongoInsert()
	if err != nil {
		log.Fatalln(err)
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", hello)

	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	e.Logger.Fatal(e.Start(port))
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "<h1>Hello wiht Golang!</h1>")
}

func mongoInsert() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoUri := fmt.Sprintf("mongodb://%s:%s@%s:%s/", os.Getenv("MONGO_ROOT_USERNAME"), os.Getenv("MONGO_ROOT_PASSWORD"), os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"))
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)

	collection := client.Database("mpostument").Collection("users")
	user := User{
		Name:    "Maksym",
		Surname: "Postument",
	}

	res, err := collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	id := res.InsertedID
	log.Printf("Insert ID, %s", id)
	return nil

}
