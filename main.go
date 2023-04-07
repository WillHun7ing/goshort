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

	"goshort/database"
)

type Link struct {
	Long  string `json:"long" xml:"long"`
	Short string `json:"short" xml:"short"`
	Visit uint32 `json:"visit" xml:"visit"`
}

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	sid, err := shortid.New(1, shortid.DefaultABC, 2342)
	shortid.SetDefault(sid)

	ctx, collection, err := mongoConnect()
	if err != nil {
		log.Fatalln(err)
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/short", func(c echo.Context) error {
		url := c.FormValue("url")
		short, err := shortid.Generate()
		if err != nil {
			log.Fatalln(err)
		}
		l := &Link{
			Long:  url,
			Short: short,
			Visit: 0,
		}
		l.insertOnDb(ctx, collection)
		return c.JSON(http.StatusOK, l)
	})
	// e.POST("/:shortUrl", func(c echo.Context) error {

	// })

	port := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	e.Logger.Fatal(e.Start(port))
}

func mongoConnect() (context.Context, *mongo.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoUri := fmt.Sprintf("mongodb://%s:%s@%s:%s/", os.Getenv("MONGO_ROOT_USERNAME"), os.Getenv("MONGO_ROOT_PASSWORD"), os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"))
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		return nil, nil, err
	}
	defer client.Disconnect(ctx)

	db := database.CreateDatabase(client, "goshort")
	c := db.CreateCollection("links")

	return ctx, c, nil
}

func (data *Link) insertOnDb(ctx context.Context, c *mongo.Collection) error {
	res, err := c.InsertOne(ctx, data)
	if err != nil {
		return err
	}
	id := res.InsertedID
	log.Printf("Insert ID, %s", id)
	return nil
}
