package main

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var health bool

// album represents data about a record album.
type album struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
	Title       string             `json:"title" bson:"title"`
	Artist      string             `json:"artist" bson:"artist"`
	Price       float64            `json:"price" bson:"price"`
	Code        string             `json:"code" bson:"code"`
	Description string             `json:"description" bson:"description"`
	Completed   bool               `json:"completed" bson:"completed"`
}

/* Used to create a singleton object of MongoDB client.
Initialized and exposed through  GetMongoClient().*/
var clientInstance *mongo.Client

//Used during creation of singleton client object in GetMongoClient().
var clientInstanceError error

//Used to execute client creation procedure only once.
var mongoOnce sync.Once

var (
	CONNECTIONSTRING = goDotEnvVariable("CONNECTIONSTRING")
	DB               = goDotEnvVariable("DB")
	ISSUES           = goDotEnvVariable("ISSUES")
	USERNAME         = goDotEnvVariable("DBUSERNAME")
	PASSWORD         = goDotEnvVariable("DBPASSWORD")
)

func goDotEnvVariable(key string) string {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}
