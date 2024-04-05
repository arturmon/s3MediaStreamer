package mongodb

import (
	"context"
	"fmt"
	"s3MediaStreamer/app/internal/config"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/pkg/logging"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoCollectionQuery interface {
	FindUser(ctx context.Context, value interface{}, columnType string) (model.User, error)
	CreateUser(ctx context.Context, user model.User) error
	DeleteUser(ctx context.Context, email string) error
	UpdateUser(ctx context.Context, email string, fields map[string]interface{}) error
	GetStoredRefreshToken(ctx context.Context, userEmail string) (string, error)
	SetStoredRefreshToken(ctx context.Context, userEmail, refreshToken string) error
	CreateTracks(ctx context.Context, list []model.Track) error
	GetTracks(ctx context.Context, offset, limit int, sortBy, sortOrder, filterArtist string) ([]model.Track, int, error)
	GetTracksByColumns(ctx context.Context, code, columns string) (*model.Track, error)
	DeleteTracks(ctx context.Context, code, columns string) error
	DeleteTracksAll(ctx context.Context) error
	UpdateTracks(ctx context.Context, track *model.Track) error
	AddTrackToPlaylist(ctx context.Context, playlistID, referenceID, referenceType string) error
	RemoveTrackFromPlaylist(ctx context.Context, playlistID, trackID string) error
	GetAllTracks(ctx context.Context) ([]model.Track, error)
	CreatePlayListName(ctx context.Context, newPlaylist model.PLayList) error
	GetPlayListByID(ctx context.Context, playlistID string) (model.PLayList, []model.Track, error)
	DeletePlaylist(ctx context.Context, playlistID string) error
	PlaylistExists(ctx context.Context, playlistID string) bool
	ClearPlayList(ctx context.Context, playlistID string) error
	UpdatePlaylistTrackOrder(ctx context.Context, playlistID string, trackOrderRequest []string) error
	GetAllTracksByPositions(ctx context.Context, playlistID string) ([]model.Track, error)
	GetAllPlayList(ctx context.Context, creatorUserID string) ([]model.PLayList, error)
	GetUserAtPlayList(ctx context.Context, playlistID string) (string, error)
	GetS3VersionByTrackID(ctx context.Context, trackID string) (string, error)
	AddS3Version(ctx context.Context, trackID, version string) error
	DeleteS3VersionByTrackID(ctx context.Context, trackID string) error
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
