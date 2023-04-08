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

// var ctx = context.TODO()

// func init() {
// 	mongoUri := fmt.Sprintf("mongodb://%s:%s@%s:%s/", os.Getenv("MONGO_ROOT_USERNAME"), os.Getenv("MONGO_ROOT_PASSWORD"), os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"))
// 	clientOptions := options.Client().ApplyURI(mongoUri)
// 	client, err := mongo.Connect(ctx, clientOptions)

// 	if err != nil {
// 		log.Fatal("Mongo init, ", err)
// 	}
// 	err = client.Ping(ctx, nil)
// 	if err != nil {
// 		log.Fatal("Mongo client pign, ", err)
// 	}
// 	collection = client.Database("goshort").Collection("links")
// }

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
	// passing bson.D{{}} matches all documents in the collection
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

	// ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	ctx := context.Background()
	// defer cancel()
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
			createLink(l)
			return c.JSON(http.StatusOK, l)
		}
		return c.JSON(http.StatusOK, result)

	})

	// e.POST("/:shortUrl", func(c echo.Context) error {
	// 	shortUrl := c.Param("shortUrl")
	// 	return c.JSON(http.StatusOK, shortUrl)
	// })

	e.GET("/", func(c echo.Context) error {
		// links, err := getAll()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		return c.JSON(http.StatusOK, &Link{})
	})

	port := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	e.Logger.Fatal(e.Start(port))
}

/*
func mongoConnect() (*context.Context, *mongo.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoUri := fmt.Sprintf("mongodb://%s:%s@%s:%s/", os.Getenv("MONGO_ROOT_USERNAME"), os.Getenv("MONGO_ROOT_PASSWORD"), os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"))
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		return nil, nil, err
	}
	defer client.Disconnect(ctx)

	// db := database.CreateDatabase(client, "goshort")
	// c := db.CreateCollection("links")

	connection := client.Database("goshort")
	c := connection.Collection("links")

	return &ctx, c, nil
}

func (data *Link) insertOnDb(ctx *context.Context, c *mongo.Collection) error {
	res, err := c.InsertOne(*ctx, data)
	if err != nil {
		return err
	}
	id := res.InsertedID
	log.Printf("Insert ID, %s", id)
	fmt.Println("😈😈😈😈 Insert Id", id)
	return nil
}
*/
