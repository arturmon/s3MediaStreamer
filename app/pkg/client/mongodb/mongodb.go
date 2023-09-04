package mongodb

import (
	"context"
	"errors"
	"fmt"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/model"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoCollectionQuery interface {
	FindCollections(name string) (*mongo.Collection, error)
	FindUserByType(value interface{}, columnType string) (model.User, error)
	CreateUser(user model.User) error
	DeleteUser(email string) error
	GetStoredRefreshToken(userEmail string) (string, error)
	SetStoredRefreshToken(userEmail, refreshToken string) error
	UpdateUserFieldsByEmail(email string, fields map[string]interface{}) error
	CreateIssue(task *model.Album) error
	CreateMany(list []model.Album) error
	GetAlbums(offset, limit int, sortBy, sortOrder, filterArtist string) ([]model.Album, int, error)
	GetIssuesByCode(code string) (model.Album, error)
	DeleteOne(code string) error
	DeleteAll() error
	UpdateIssue(album *model.Album) error
}

type MongoOperations interface {
	MongoCollectionQuery
}

type MongoClient struct {
	Client *mongo.Client
	Cfg    *config.Config
}

func (c *MongoClient) Connect(_ *logging.Logger) error {
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

func (c *MongoClient) CreateIssue(album *model.Album) error {
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

func (c *MongoClient) CreateMany(issues []model.Album) error {
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

func (c *MongoClient) GetIssuesByCode(code string) (model.Album, error) {
	result := model.Album{}
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

func (c *MongoClient) GetAlbums(offset, limit int, sortBy, sortOrder, filterArtist string) ([]model.Album, int, error) {
	if offset < 1 || limit < 1 {
		return nil, 0, errors.New("invalid pagination parameters")
	}

	filter := bson.D{}
	if filterArtist != "" {
		filter = append(filter, bson.E{Key: "artist", Value: bson.M{"$regex": filterArtist, "$options": "i"}})
	}

	var albums []model.Album

	collection, err := c.FindCollections(config.CollectionAlbum)
	if err != nil {
		return albums, 0, err
	}

	findOptions := options.Find()
	findOptions.SetSkip(int64((offset - 1) * limit))
	findOptions.SetLimit(int64(limit))

	sortOptions := options.Find()
	sortOrderValue := 1
	if sortOrder == "desc" {
		sortOrderValue = -1
	}
	sortOptions.SetSort(bson.D{{Key: sortBy, Value: sortOrderValue}})
	findOptions.Sort = sortOptions.Sort

	cur, findError := collection.Find(context.TODO(), filter, findOptions)
	if findError != nil {
		return albums, 0, findError
	}

	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var album model.Album
		decodeErr := cur.Decode(&album)
		if decodeErr != nil {
			return albums, 0, decodeErr
		}
		albums = append(albums, album)
	}

	totalCount, countErr := collection.CountDocuments(context.TODO(), filter)
	if countErr != nil {
		return albums, 0, countErr
	}

	if len(albums) == 0 {
		return albums, 0, mongo.ErrNoDocuments
	}
	return albums, int(totalCount), nil
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

func (c *MongoClient) CreateUser(user model.User) error {
	collection, err := c.FindCollections(config.CollectionUser)
	if err != nil {
		return err
	}

	// Поиск пользователя с помощью email.
	existingUser, err := c.FindUserByType(user.Email, "email")
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

func (c *MongoClient) FindUserByType(value interface{}, columnType string) (model.User, error) {
	result := model.User{}
	// Define filter query for fetching a specific document from the collection.
	filter := bson.D{{Key: columnType, Value: value}}
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

// GetStoredRefreshToken retrieves the refresh token for a user by email.
func (c *MongoClient) GetStoredRefreshToken(userEmail string) (string, error) {
	result := model.User{}
	collection, err := c.FindCollections(config.CollectionUser)
	if err != nil {
		return "", err
	}
	filter := bson.M{"email": userEmail}
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", fmt.Errorf("user with email '%s' not found", userEmail)
	}
	return result.RefreshToken, nil
}

// SetStoredRefreshToken sets or updates the refresh token for a user by email.
func (c *MongoClient) SetStoredRefreshToken(userEmail, refreshToken string) error {
	collection, err := c.FindCollections(config.CollectionUser)
	if err != nil {
		return err
	}
	filter := bson.M{"email": userEmail}
	update := bson.M{"$set": bson.M{"refreshtoken": refreshToken}}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (c *MongoClient) UpdateIssue(album *model.Album) error {
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

func (c *MongoClient) UpdateUserFieldsByEmail(email string, fields map[string]interface{}) error {
	// Define the filter to find the user by email.
	filter := bson.M{"email": email}

	// Define the update statement based on the provided fields.
	update := bson.M{}
	for key, value := range fields {
		update[key] = value
	}

	// Create an options instance to enable upsert (insert if document not found).
	options := options.Update().SetUpsert(false)

	// Perform the update operation.
	collection, err := c.FindCollections(config.CollectionAlbum)
	if err != nil {
		return err
	}
	_, err = collection.UpdateOne(context.TODO(), filter, bson.M{"$set": update}, options)
	if err != nil {
		return err
	}

	return nil
}
