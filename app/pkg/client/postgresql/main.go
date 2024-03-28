package postgresql

import (
	"context"
	"fmt"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/pkg/logging"
	"time"

	"github.com/jackc/pgx/v5"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCollectionQuery interface {
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
	AddTrackToPlaylist(ctx context.Context, playlistID, trackID string) error
	RemoveTrackFromPlaylist(ctx context.Context, playlistID, trackID string) error
	GetAllTracks(ctx context.Context) ([]model.Track, error)
	CreatePlayListName(ctx context.Context, newPlaylist model.PLayList) error
	GetPlayListByID(ctx context.Context, playlistID string) (model.PLayList, []model.Track, error)
	DeletePlaylist(ctx context.Context, playlistID string) error
	PlaylistExists(ctx context.Context, playlistID string) bool
	ClearPlayList(ctx context.Context, playlistID string) error
	UpdatePlaylistTrackOrder(ctx context.Context, playlistID string, trackOrderRequest []string) error
	GetAllTracksByPositions(ctx context.Context, playlistID string) ([]model.Track, error)
	GetAllPlayList(ctx context.Context) ([]model.PLayList, error)
}

type PostgresOperations interface {
	PostgresCollectionQuery
}

type PgClient struct {
	Pool             *pgxpool.Pool
	ConnectionString string
}

func (c *PgClient) GetConnectionString() string {
	return c.ConnectionString
}

func (c *PgClient) Begin(ctx context.Context) (pgx.Tx, error) {
	return c.Pool.Begin(ctx)
}

func (c *PgClient) FindCollections(name string) (*mongo.Collection, error) {
	return nil, fmt.Errorf("FindCollections is not supported for PostgreSQL, %s not finded", name)
}

func DoWithAttempts(fn func() error, maxAttempts int, delay time.Duration) error {
	var err error

	for maxAttempts > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			maxAttempts--

			continue
		}

		return nil
	}

	return err
}

func (c *PgClient) Connect(_ *logging.Logger) error {
	if c.Pool != nil {
		conn, connErr := c.Pool.Acquire(context.Background())
		if connErr != nil {
			return connErr
		}
		defer conn.Release()
		if pingErr := conn.Conn().Ping(context.Background()); pingErr != nil {
			return pingErr
		}
	} else {
		return fmt.Errorf("pgx pool is not initialized")
	}
	return nil
}

func (c *PgClient) Ping(ctx context.Context) error {
	if c.Pool != nil {
		conn, connErr := c.Pool.Acquire(ctx)
		if connErr != nil {
			return connErr
		}
		defer conn.Release()
		pingErr := conn.Conn().Ping(ctx)
		if pingErr != nil {
			return pingErr
		}
	} else {
		return fmt.Errorf("pgx pool is not initialized")
	}
	return nil
}

func (c *PgClient) Close(_ context.Context) error {
	if c.Pool != nil {
		c.Pool.Close()
		c.Pool = nil
	}
	return nil
}
