package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/teris-io/shortid"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	Name  string `json:"name" xml:"name"`
	Email string `json:"email" xml:"email"`
	Long  string `json:"long" xml:"long"`
	Short string `json:"short" xml:"short"`
}

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	sid, err := shortid.New(1, shortid.DefaultABC, 2342)
	shortid.SetDefault(sid)

	err = mongoInsert()
	if err != nil {
		log.Fatalln(err)
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/short", makeShortenedLink)

	port := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	e.Logger.Fatal(e.Start(port))
}

func makeShortenedLink(c echo.Context) error {
	url := c.FormValue("url")
	short, err := shortid.Generate()
	if err != nil {
		log.Fatalln(err)
	}
	u := &User{
		Name:  "Mohammadreza",
		Email: "h9edev@gmail.com",
		Long:  url,
		Short: short,
	}
	return c.JSON(http.StatusOK, u)
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
		Name:  "Maksym",
		Email: "h9edev@gmail.com",
		Long:  "",
		Short: "",
	}

	res, err := collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	id := res.InsertedID
	log.Printf("Insert ID, %s", id)
	return nil

}
