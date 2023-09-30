package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hajbabaeim/goshort/internal/database"

	"github.com/teris-io/shortid"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/go-redis/redis"
)

type Link struct {
	Long  string `json:"long" bson:"long" xml:"long"`
	Short string `json:"short" bson:"short" xml:"short"`
	Visit uint32 `json:"visit" bson:"visit" xml:"visit"`
}

var collection *mongo.Collection

func createLink(ctx context.Context, link *Link) error {
	_, err := collection.InsertOne(ctx, link)
	if err != nil {
		log.Fatalln("Insert error: ", err)
	}
	addToCache(link.Long, link)
	return err
}

func checkCachedValues(key string, link *Link) (bool, error) {
	isCached, err := getFromCache(key, link)
	if err != nil {
		addToCache(link.Long, link)
		return false, err
	}
	return isCached, nil
}

func FetchItemFromCacheOrMongo(ctx context.Context, longUrl string, result *Link) error {
	filter := bson.D{{"long", longUrl}}
	isCached, _ := checkCachedValues(longUrl, result)
	if isCached {
		return nil
	}
	err := collection.FindOne(ctx, filter).Decode(result)
	if err != nil {
		short, err := shortid.Generate()
		if err != nil {
			log.Fatalln(err)
		}
		l := &Link{
			Long:  longUrl,
			Short: short,
			Visit: 0,
		}
		createLink(ctx, l)
		*result = *l
	}
	return nil
}

func FindItemOnMongo(ctx context.Context, shortUrl string, link *Link) error {
	filter := bson.D{{"short", shortUrl}}
	err := collection.FindOne(ctx, filter).Decode(link)
	if err != nil {
		return err
	}
	return nil
}

func incrementVisit(ctx context.Context, shortUrl string, link *Link) error {
	var result Link
	FindItemOnMongo(ctx, shortUrl, &result)
	filter := bson.D{{"short", shortUrl}}
	update := bson.D{{"$set", bson.D{{"visit", result.Visit + 1}}}}
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Update value had an issue")
	}
	FindItemOnMongo(ctx, shortUrl, link)
	addToCache(link.Long, link)
	return nil
}

func getAll(ctx context.Context) ([]*Link, error) {
	filter := bson.D{{}}
	return filterLinks(ctx, filter)
}

func filterLinks(ctx context.Context, filter interface{}) ([]*Link, error) {
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

func addToCache(key string, link *Link) error {
	var redisUri string
	if os.Getenv("ENV") == "docker" {
		redisUri = fmt.Sprintf("redis:%s", os.Getenv("REDIS_PORT"))
	} else {
		redisUri = fmt.Sprintf("localhost:%s", os.Getenv("REDIS_PORT"))
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisUri,
		// Password: "123456",
		DB: 0,
	})

	jsonString, err := json.Marshal(link)
	if err != nil {
		log.Fatalf("Error while marshaling data, %s", err)
		return err
	}
	err = redisClient.Set(key, jsonString, 24*time.Hour).Err()
	if err != nil {
		log.Fatalf("Error while storing data to redis, %s", err)
		return err
	}

	return nil
}

func getFromCache(key string, link *Link) (bool, error) {
	var redisUri string
	if os.Getenv("ENV") == "docker" {
		redisUri = fmt.Sprintf("redis:%s", os.Getenv("REDIS_PORT"))
	} else {
		redisUri = fmt.Sprintf("localhost:%s", os.Getenv("REDIS_PORT"))
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisUri,
		// Password: "123456",
		DB: 0,
	})

	cachedLink, err := redisClient.Get(key).Bytes()
	if err != nil {
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
	if err != nil {
		log.Fatalf("Error creating shortid generator: %v", err)
	}
	shortid.SetDefault(sid)

	ctx := context.Background()
	mongoUri := getMongoURI()
	client := database.CreateClient(mongoUri)
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatalf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	goshortDB := database.CreateDatabase(client, "goshort")
	collection = goshortDB.CreateCollection("links")

	// collection = client.Database("goshort").Collection("links")

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/short", func(c echo.Context) error {
		url := c.FormValue("url")
		var result Link
		FetchItemFromCacheOrMongo(ctx, url, &result)
		return c.JSON(http.StatusOK, result)
	})
	e.POST("/:shortUrl", func(c echo.Context) error {
		shortUrl := c.Param("shortUrl")
		var result Link
		incrementVisit(ctx, shortUrl, &result)
		return c.JSON(http.StatusOK, result)
	})
	e.GET("/", func(c echo.Context) error {
		links, err := getAll(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Cannot find anything")
		}
		return c.JSON(http.StatusOK, links)
	})

	port := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	e.Logger.Fatal(e.Start(port))
}

func getMongoURI() string {
	if os.Getenv("ENV") == "docker" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%s/", os.Getenv("MONGO_ROOT_USERNAME"), os.Getenv("MONGO_ROOT_PASSWORD"), os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"))
	}
	return fmt.Sprintf("mongodb://127.0.0.1:%s/", os.Getenv("MONGO_PORT"))
}
