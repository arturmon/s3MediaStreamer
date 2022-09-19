package mongodb

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	//"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	//"skeleton-golange-application/app/internal/app"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/model"
	"skeleton-golange-application/app/pkg/monitoring"
	"sync"
)

// NewMongoConfig creates new pg config instance
func NewMongoConfig(username, password, host, port, database, collections string) *model.StorageConfig {
	return &model.StorageConfig{
		Host:        host,
		Port:        port,
		Database:    database,
		Collections: collections,
		Username:    username,
		Password:    password,
	}
}

//Used to execute client creation procedure only once.
var mongoOnce sync.Once

//GetMongoClient - Return mongodb connection to work with
func GetMongoClient(cfg *model.StorageConfig) (clientInstance *mongo.Client, clientInstanceError error) {

	//Perform connection creation operation only once.
	connectionString := fmt.Sprintf("mongodb://%s:%s", cfg.Host, cfg.Port)

	mongoOnce.Do(func() {
		credential := options.Credential{
			Username: cfg.Username,
			Password: cfg.Password,
		}
		// Set client options
		clientOptions := options.Client().ApplyURI(connectionString).SetAuth(credential)
		// Connect to MongoDB
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			clientInstanceError = err
		}
		// Check the connection
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			clientInstanceError = err
		}
		clientInstance = client
	})
	if clientInstanceError != nil {
		log.Error(clientInstanceError)
		// prometheuse
		monitoring.GetAlbumsErrorConnectMongodbTotal.Inc()

		config.AppHealth = false
	} else {
		log.Info("Successfully connected and pinged ", connectionString)
		// prometheuse
		monitoring.CountGetAlbumsConnectMongodbTotal.Inc()

		config.AppHealth = true
	}
	return clientInstance, clientInstanceError
}

func findCollections(cfg *config.Config, client *mongo.Client, useCollections string) (collection *mongo.Collection, err error) {
	/*
		client, err = GetMongoClient(cfg)
		if err != nil {
			return nil, err
		}
	*/
	switch useCollections {
	case "album":
		collection = client.Database(cfg.Storage.MongoDB.Database).Collection(cfg.Storage.MongoDB.Collections)
	case "user":
		collection = client.Database(cfg.Storage.MongoDB.Database).Collection(cfg.Storage.MongoDB.CollectionsUsers)
	}
	return collection, nil
}

/*
------------------------------------------------------------------------------
*/

//CreateIssue - Insert a new document in the collection.
func CreateIssue(cfg *config.Config, client *mongo.Client, task config.Album) error {
	collection, err := findCollections(cfg, client, config.COLLECTION_ALBUM)
	if err != nil {
		return err
	}
	//Perform InsertOne operation & validate against the error.
	_, err = collection.InsertOne(context.TODO(), task)
	if err != nil {
		return err
	}
	//Return success without any error.
	return nil
}

//CreateMany - Insert multiple documents at once in the collection.
func CreateMany(cfg *config.Config, client *mongo.Client, list []config.Album) error {
	//Map struct slice to interface slice as InsertMany accepts interface slice as parameter
	insertableList := make([]interface{}, len(list))
	for i, v := range list {
		insertableList[i] = v
	}
	collection, err := findCollections(cfg, client, config.COLLECTION_ALBUM)
	if err != nil {
		return err
	}
	//Perform InsertMany operation & validate against the error.
	_, err = collection.InsertMany(context.TODO(), insertableList)
	if err != nil {
		return err
	}
	//Return success without any error.
	return nil
}

//GetIssuesByCode - Get All issues for collection
func GetIssuesByCode(cfg *config.Config, client *mongo.Client, code string) (config.Album, error) {
	result := config.Album{}
	//Define filter query for fetching specific document from collection
	filter := bson.D{primitive.E{Key: "code", Value: code}}
	collection, err := findCollections(cfg, client, config.COLLECTION_ALBUM)
	if err != nil {
		return result, err
	}
	//Perform FindOne operation & validate against the error.
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return result, err
	}
	//Return result without any error.
	return result, nil
}

func GetAllIssues(cfg *config.Config, client *mongo.Client) ([]config.Album, error) {
	//Define filter query for fetching specific document from collection
	filter := bson.D{{}} //bson.D{{}} specifies 'all documents'
	var issues []config.Album

	collection, err := findCollections(cfg, client, config.COLLECTION_ALBUM)
	if err != nil {
		return issues, err
	}
	//Perform Find operation & validate against the error.
	cur, findError := collection.Find(context.TODO(), filter)
	if findError != nil {
		return issues, findError
	}
	//Map result to slice
	for cur.Next(context.TODO()) {
		var t config.Album
		err := cur.Decode(&t)
		if err != nil {
			return issues, err
		}
		issues = append(issues, t)
	}
	// once exhausted, close the cursor
	cur.Close(context.TODO())
	if len(issues) == 0 {
		return issues, mongo.ErrNoDocuments
	}
	return issues, nil
}

//DeleteOne - Delete One document
func DeleteOne(cfg *config.Config, client *mongo.Client, code string) error {
	//Define filter query for fetching specific document from collection
	filter := bson.D{primitive.E{Key: "code", Value: code}}
	collection, err := findCollections(cfg, client, config.COLLECTION_ALBUM)
	if err != nil {
		return err
	}
	//Perform DeleteOne operation & validate against the error.
	_, err = collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	//Return success without any error.
	return nil
}

//DeleteAll - Delte All dockument
func DeleteAll(cfg *config.Config, client *mongo.Client) error {
	//Define filter query for fetching specific document from collection
	selector := bson.D{{}} // bson.D{{}} specifies 'all documents'
	collection, err := findCollections(cfg, client, config.COLLECTION_ALBUM)
	if err != nil {
		return err
	}
	//Perform DeleteMany operation & validate against the error.
	_, err = collection.DeleteMany(context.TODO(), selector)
	if err != nil {
		return err
	}
	//Return success without any error.
	return nil
}

//PrintList - Print list of issues on console
func PrintList(issues []config.Album) {
	for i, v := range issues {
		if v.Completed {
			fmt.Printf("INFO: Completed %d: %f    %s\n", i+1, v.Price, v.Title)
		} else {
			fmt.Printf("INFO: No Completed %d: %f    %s\n", i+1, v.Price, v.Title)
		}
	}
}

// MarkCompleted - MarkCompleted
func MarkCompleted(cfg *config.Config, client *mongo.Client, code string) error {
	//Define filter query for fetching specific document from collection
	filter := bson.D{primitive.E{Key: "code", Value: code}}

	//Define updater for to specifiy change to be updated.
	updater := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "completed", Value: true},
	}}}

	collection, err := findCollections(cfg, client, config.COLLECTION_ALBUM)
	if err != nil {
		return err
	}
	//Perform UpdateOne operation & validate against the error.
	_, err = collection.UpdateOne(context.TODO(), filter, updater)
	if err != nil {
		return err
	}
	//Return success without any error.
	return nil
}

func CreateUser(cfg *config.Config, client *mongo.Client, user config.User) error {

	collection, err := findCollections(cfg, client, config.COLLECTION_USER)
	if err != nil {
		return err
	}
	//Perform InsertOne operation & validate against the error.
	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}
	//Return success without any error.
	return nil

	// TODO Создать функцию поискать пользователя
	//_, err := mongodb.GetIssuesByUser(a.cfg, a.mongoClient, user.Code)
	/*
			if err == mongo.ErrNoDocuments {
				// TODO Создать функцию создания пользователя
				mongodb.CreateUser(a.cfg, a.mongoClient, user)
				c.IndentedJSON(http.StatusCreated, user)
			}


		//return c.JSON(user)
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Not Create User"})

	*/
	return nil
}
func FindUserToEmail(cfg *config.Config, client *mongo.Client, email string) (config.User, error) {
	result := config.User{}
	//Define filter query for fetching specific document from collection
	filter := bson.D{primitive.E{Key: "email", Value: email}}
	collection, err := findCollections(cfg, client, config.COLLECTION_USER)
	if err != nil {
		return result, err
	}

	//Perform FindOne operation & validate against the error.
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return result, err
	}
	//Return result without any error.
	return result, nil
}

//database.DB.Where("email = ?", data["email"]).First(&user)
//database.DB.Where("id = ?", claims.Issuer).First(&user)

func User(cfg *config.Config, client *mongo.Client, claims *jwt.StandardClaims) error {

	//database.DB.Where("id = ?", claims.Issuer).First(&user)
	// TODO Создать функцию поискать пользователя
	//_, err := mongodb.GetIssuesByUser(a.cfg, a.mongoClient, user.Code)
	/*
			if err == mongo.ErrNoDocuments {
				// TODO Создать функцию создания пользователя
				mongodb.CreateUser(a.cfg, a.mongoClient, user)
				c.IndentedJSON(http.StatusCreated, user)
			}


		//return c.JSON(user)
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Not Create User"})

	*/
	return nil
}
