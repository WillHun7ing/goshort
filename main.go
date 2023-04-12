package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/teris-io/shortid"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/go-redis/redis"
)

type Link struct {
	Long  string `json:"long" bson:"long" xml:"long"`
	Short string `json:"short" bson:"short" xml:"short"`
	Visit uint32 `json:"visit" bson:"visit" xml:"visit"`
}

var collection *mongo.Collection

var ctx context.Context
var cancel context.CancelFunc

func createLink(link *Link) error {
	_, err := collection.InsertOne(ctx, link)
	if err != nil {
		log.Fatalln("Insert error: ", err)
	}
	return err
}

func incrementVisit(filter bson.D, link *Link) error {
	update := bson.D{{"$set", bson.D{{"visit", link.Visit + 1}}}}
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Update value went with a problem")
	}
	return nil
}

func checkCachedValues(shortUrl string) error {
	// var respMessage map[string]interface{}
	var l Link
	isCached, err := getFromCache(shortUrl, &l)
	if err != nil {
		return err
	}
	if isCached == true {
		// l["_source"] = "Redis"
	}
	return nil
}

func findLinkWithShortUrl(filter bson.D, result *Link) error {
	err := collection.FindOne(ctx, filter).Decode(result)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Please provide valid shortened url")
	}
	return nil
}

func getAll() ([]*Link, error) {
	filter := bson.D{{}}
	return filterLinks(filter)
}

func filterLinks(filter interface{}) ([]*Link, error) {
	var links []*Link
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return links, err
	}
	for cursor.Next(ctx) {
		var l Link
		err := cursor.Decode(&l)
		if err != nil {
			return links, err
		}
		links = append(links, &l)
	}
	if err := cursor.Err(); err != nil {
		return links, err
	}
	cursor.Close(ctx)
	if len(links) == 0 {
		return links, mongo.ErrNoDocuments
	}

	return links, nil
}

func addToCache(cacheName string, link *Link) error {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		// Password: "123456",
		DB: 0,
	})

	jsonString, err := json.Marshal(link)
	if err != nil {
		log.Fatalf("Error while marshaling data, %s", err)
		return err
	}
	err = redisClient.Set(cacheName, jsonString, 24*time.Hour).Err()
	if err != nil {
		log.Fatalf("Error while storing data to redis, %s", err)
		return err
	}

	return nil
}

func getFromCache(cacheName string, link *Link) (bool, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		// Password: "123456",
		DB: 0,
	})

	cachedLink, err := redisClient.Get(cacheName).Bytes()
	if err != nil {
		log.Fatalf("Error while retrieving data from redis, %s", err)
		return false, nil
	}
	err = json.Unmarshal(cachedLink, &link)
	if err != nil {
		log.Fatalf("Error while unmarshaling data, %s", err)
		return false, nil
	}

	return true, nil
}

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	sid, err := shortid.New(1, shortid.DefaultABC, 2342)
	shortid.SetDefault(sid)

	ctx := context.Background()
	// mongoUri := fmt.Sprintf("mongodb://%s:%s@%s:%s/", os.Getenv("MONGO_ROOT_USERNAME"), os.Getenv("MONGO_ROOT_PASSWORD"), os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"))
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:27017/"))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	collection = client.Database("goshort").Collection("links")

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/short", func(c echo.Context) error {
		url := c.FormValue("url")
		var result Link
		err := collection.FindOne(ctx, bson.D{{"long", url}}).Decode(&result)
		if err != nil {
			short, err := shortid.Generate()
			if err != nil {
				log.Fatalln(err)
			}
			l := &Link{
				Long:  url,
				Short: short,
				Visit: 0,
			}
			// var respMessage map[string]interface{}
			// respMessage[short] = l
			addToCache(short, l)
			createLink(l)
			return c.JSON(http.StatusOK, l)
		}
		return c.JSON(http.StatusOK, result)

	})

	e.POST("/:shortUrl", func(c echo.Context) error {
		shortUrl := c.Param("shortUrl")
		var result Link
		filter := bson.D{{"short", shortUrl}}
		checkCachedValues(shortUrl)
		findLinkWithShortUrl(filter, &result)
		incrementVisit(filter, &result)
		findLinkWithShortUrl(filter, &result)
		return c.JSON(http.StatusOK, result)
	})

	e.GET("/", func(c echo.Context) error {
		links, err := getAll()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Cannot find anything")
		}
		return c.JSON(http.StatusOK, links)
	})

	port := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	e.Logger.Fatal(e.Start(port))
}
