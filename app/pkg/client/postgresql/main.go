package postgresql

import (
	"context"
	"fmt"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/model"
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
	CreateAlbums(list []model.Album) error
	GetAlbums(offset, limit int, sortBy, sortOrder, filterArtist string) ([]model.Album, int, error)
	GetAlbumsByCode(code string) (model.Album, error)
	DeleteAlbums(code string) error
	DeleteAlbumsAll() error
	UpdateAlbums(album *model.Album) error
	GetAlbumsForLearn() ([]model.Album, error)
	CreateTops(list []model.Tops) error
	CleanupRecords(retentionPeriod time.Duration) error
	GetAllAlbums() ([]model.Album, error)
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
