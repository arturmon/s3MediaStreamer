package mongodb

import (
	"context"
	"fmt"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/model"
	"skeleton-golange-application/app/pkg/logging"
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
	CreateTracks(list []model.Track) error
	GetTracks(offset, limit int, sortBy, sortOrder, filterArtist string) ([]model.Track, int, error)
	GetTracksByColumns(code, columns string) (*model.Track, error)
	DeleteTracks(code, columns string) error
	DeleteTracksAll() error
	UpdateTracks(track *model.Track) error
	AddTrackToPlaylist(playlistID, trackID string) error
	RemoveTrackFromPlaylist(playlistID, trackID string) error
	GetTracksForLearn() ([]model.Track, error)
	CreateTops(list []model.Tops) error
	CleanupRecords(retentionPeriod time.Duration) error
	GetAllTracks() ([]model.Track, error)
	CreatePlayListName(newPlaylist model.PLayList) error
	GetPlayListByID(playlistID string) (model.PLayList, []model.Track, error)
	DeletePlaylist(playlistID string) error
	PlaylistExists(playlistID string) bool
	ClearPlayList(playlistID string) error
	UpdatePlaylistTrackOrder(playlistID string, trackOrderRequest []string) error
	GetAllTracksByPositions(playlistID string) ([]model.Track, error)
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
	case "track":
		return c.Client.Database(c.Cfg.Storage.Database).Collection(c.Cfg.Storage.Collections), nil
	case "user":
		return c.Client.Database(c.Cfg.Storage.Database).Collection(c.Cfg.Storage.CollectionsUsers), nil
	default:
		return nil, fmt.Errorf("unsupported collection type: %s", useCollections)
	}
}
