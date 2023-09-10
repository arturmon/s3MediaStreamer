package mongodb

import (
	"context"
	"fmt"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/model"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoCollectionQuery interface {
	FindUser(value interface{}, columnType string) (model.User, error)
	CreateUser(user model.User) error
	DeleteUser(email string) error
	UpdateUser(email string, fields map[string]interface{}) error
	GetStoredRefreshToken(userEmail string) (string, error)
	SetStoredRefreshToken(userEmail, refreshToken string) error
	CreateAlbums(list []model.Album) error
	GetAlbums(offset, limit int, sortBy, sortOrder, filterArtist string) ([]model.Album, int, error)
	GetAlbumsByCode(code string) (model.Album, error)
	DeleteAlbums(code string) error
	DeleteAlbumsAll() error
	UpdateAlbums(album *model.Album) error
	GetAlbumsForLearn() ([]model.Album, error)
	CreateTops(list []model.Tops) error
	CleanupRecords(retentionPeriod time.Duration) error
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
