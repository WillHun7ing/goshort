package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/teris-io/shortid"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func getAll() ([]*Link, error) {
	filter := bson.D{{}}
	return filterTasks(filter)
}

func filterTasks(filter interface{}) ([]*Link, error) {
	// A slice of tasks for storing the decoded documents
	var tasks []*Link

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return tasks, err
	}

	for cur.Next(ctx) {
		var t Link
		err := cur.Decode(&t)
		if err != nil {
			return tasks, err
		}

		tasks = append(tasks, &t)
	}

	if err := cur.Err(); err != nil {
		return tasks, err
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	if len(tasks) == 0 {
		return tasks, mongo.ErrNoDocuments
	}

	return tasks, nil
}

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	sid, err := shortid.New(1, shortid.DefaultABC, 2342)
	shortid.SetDefault(sid)

	ctx := context.Background()
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
			createLink(l)
			return c.JSON(http.StatusOK, l)
		}
		return c.JSON(http.StatusOK, result)

	})

	e.POST("/:shortUrl", func(c echo.Context) error {
		shortUrl := c.Param("shortUrl")
		var result Link
		err := collection.FindOne(ctx, bson.D{{"short", shortUrl}}).Decode(&result)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Please provide valid shortened url")
		}
		return c.JSON(http.StatusOK, result)
	})

	e.GET("/", func(c echo.Context) error {
		cursor, err := collection.Find(context.TODO(), bson.D{})
		if err != nil {
			fmt.Println("Finding all documtes err: ", err)
			defer cursor.Close(ctx)
		} else {
			var result bson.M
			for cursor.Next(ctx) {
				err := cursor.Decode(&result)
				if err != nil {
					fmt.Println("Get next document error: ", err)
				} else {
					fmt.Println("The value of link: ", result)
				}
			}
			return c.JSON(http.StatusOK, result)
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot find anything")
	})

	port := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	e.Logger.Fatal(e.Start(port))
}
