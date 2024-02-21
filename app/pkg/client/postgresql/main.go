package postgresql

import (
	"context"
	"fmt"
	"skeleton-golange-application/app/model"
	"skeleton-golange-application/app/pkg/logging"
	"time"

	"github.com/jackc/pgx/v4"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresCollectionQuery interface {
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
	CleanSessions() error
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
