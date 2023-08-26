package mongodb

import (
	"context"
	"errors"
	"fmt"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/logging"

	"go.mongodb.org/mongo-driver/mongo/options"

	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoCollectionQuery interface {
	FindCollections(name string) (*mongo.Collection, error)
	FindUserToEmail(email string) (config.User, error)
	CreateUser(user config.User) error
	DeleteUser(email string) error
	CreateIssue(task *config.Album) error
	CreateMany(list []config.Album) error
	GetAllIssues() ([]config.Album, error)
	GetIssuesByCode(code string) (config.Album, error)
	DeleteOne(code string) error
	DeleteAll() error
	UpdateIssue(album *config.Album) error
}

type MongoOperations interface {
	MongoCollectionQuery
}

type MongoClient struct {
	Client *mongo.Client
	Cfg    *config.Config
}

func (c *MongoClient) Connect(logger *logging.Logger) error {
	uri := "mongodb://" + c.Cfg.Storage.Host + ":" + c.Cfg.Storage.Port
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}

	c.Client = client
	return nil
}

func (c *MongoClient) DeleteUser(email string) error {
	return fmt.Errorf("DeleteUser is not supported for MongoDB, '%s' not deleted", email)
}

func (c *MongoClient) Ping(ctx context.Context) error {
	if c.Client != nil {
		return c.Client.Ping(ctx, readpref.Primary())
	}
	return fmt.Errorf("mongo client is not initialized")
}

func (c *MongoClient) Close(ctx context.Context) error {
	if c.Client != nil {
		err := c.Client.Disconnect(ctx)
		if err != nil {
			return err
		}
		c.Client = nil
	}
	return nil
}

func (c *MongoClient) FindCollections(useCollections string) (*mongo.Collection, error) {
	switch useCollections {
	case "album":
		return c.Client.Database(c.Cfg.Storage.Database).Collection(c.Cfg.Storage.Collections), nil
	case "user":
		return c.Client.Database(c.Cfg.Storage.Database).Collection(c.Cfg.Storage.CollectionsUsers), nil
	default:
		return nil, fmt.Errorf("unsupported collection type: %s", useCollections)
	}
}

func (c *MongoClient) CreateIssue(album *config.Album) error {
	collection, err := c.FindCollections(config.CollectionAlbum)
	if err != nil {
		return err
	}
	_, err = collection.InsertOne(context.TODO(), album)
	if err != nil {
		return err
	}
	return nil
}

func (c *MongoClient) CreateMany(issues []config.Album) error {
	insertableList := make([]interface{}, len(issues))
	for i := range issues {
		v := &issues[i] // Use a pointer to the current issue.
		insertableList[i] = v
		/*
			if v.Completed {
				log.Infof("INFO: Completed %d: %f    %s\n", i+1, v.Price, v.Title)
			} else {
				log.Infof("INFO: No Completed %d: %f    %s\n", i+1, v.Price, v.Title)
			}

		*/
	}

	collection, err := c.FindCollections(config.CollectionAlbum)
	if err != nil {
		return err
	}
	_, err = collection.InsertMany(context.TODO(), insertableList)
	if err != nil {
		return err
	}
	return nil
}

func (c *MongoClient) GetIssuesByCode(code string) (config.Album, error) {
	result := config.Album{}
	filter := bson.D{{Key: "code", Value: code}} // Fix the linting issue here
	collection, err := c.FindCollections(config.CollectionAlbum)
	if err != nil {
		return result, err
	}
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (c *MongoClient) GetAllIssues() ([]config.Album, error) {
	filter := bson.D{{}}
	var issues []config.Album

	collection, err := c.FindCollections(config.CollectionAlbum)
	if err != nil {
		return issues, err
	}
	cur, findError := collection.Find(context.TODO(), filter)
	if findError != nil {
		return issues, findError
	}
	for cur.Next(context.TODO()) {
		var t config.Album
		decodeErr := cur.Decode(&t)
		if decodeErr != nil {
			return issues, decodeErr
		}
		issues = append(issues, t)
	}
	cur.Close(context.TODO())
	if len(issues) == 0 {
		return issues, mongo.ErrNoDocuments
	}
	return issues, nil
}

func (c *MongoClient) DeleteOne(code string) error {
	filter := bson.D{primitive.E{Key: "code", Value: code}}
	collection, err := c.FindCollections(config.CollectionAlbum)
	if err != nil {
		return err
	}
	_, err = collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func (c *MongoClient) DeleteAll() error {
	selector := bson.D{{}}
	collection, err := c.FindCollections(config.CollectionAlbum)
	if err != nil {
		return err
	}
	_, err = collection.DeleteMany(context.TODO(), selector)
	if err != nil {
		return err
	}
	return nil
}

// PrintList - Print list of issues on console.
func PrintList(issues []config.Album) {
	for i := range issues {
		v := &issues[i] // Use a pointer to the current issue.
		log.Infof("Completed %d: %f    %s", i+1, v.Price, v.Title)
		/*
			if v.Completed {
				log.Infof("Completed %d: %f    %s", i+1, v.Price, v.Title)
			} else {
				log.Infof("No Completed %d: %f    %s", i+1, v.Price, v.Title)
			}

		*/
	}
}

func (c *MongoClient) CreateUser(user config.User) error {
	collection, err := c.FindCollections(config.CollectionUser)
	if err != nil {
		return err
	}

	// Поиск пользователя с помощью email.
	existingUser, err := c.FindUserToEmail(user.Email)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}

		// Создание нового пользователя
		_, err = collection.InsertOne(context.TODO(), user)
		if err != nil {
			return err
		}

		return nil
	}

	// Пользователь с таким email уже существует.
	return fmt.Errorf("user with email '%s' already exists", existingUser.Email)
}

func (c *MongoClient) FindUserToEmail(email string) (config.User, error) {
	result := config.User{}
	// Define filter query for fetching a specific document from the collection.
	filter := bson.D{{Key: "email", Value: email}} // Fix the linting issue here
	collection, err := c.FindCollections(config.CollectionUser)
	if err != nil {
		return result, err
	}
	// Perform FindOne operation and validate against errors.
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return result, err
	}
	// Return the result without any error.
	return result, nil
}

func (c *MongoClient) UpdateIssue(album *config.Album) error {
	collection, err := c.FindCollections(config.CollectionAlbum)
	if err != nil {
		return err
	}

	// Define the filter to find the album by its code.
	filter := bson.D{{Key: "code", Value: album.Code}}

	// Define the update fields using the $set operator to update only the specified fields.
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "title", Value: album.Title},
			{Key: "artist", Value: album.Artist},
			{Key: "price", Value: album.Price},
			{Key: "description", Value: album.Description},
			{Key: "sender", Value: album.Sender},
		}},
		{Key: "$currentDate", Value: bson.D{
			{Key: "updatedat", Value: true}, // Set the "updatedat" field to the current date.
		}},
	}

	// Perform the update operation.
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}
