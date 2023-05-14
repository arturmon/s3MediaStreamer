package model

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/mongodb"
	"skeleton-golange-application/app/pkg/client/postgresql"
	"time"
)

type StorageConfig struct {
	Type             string
	Host             string
	Port             string
	Database         string
	Collections      string
	CollectionsUsers string
	Username         string
	Password         string
	client           *mongo.Client
	pool             *pgxpool.Pool
}

const (
	// maxAttempts is the maximum number of attempts to connect to the database.
	maxAttempts = 10
	// maxDelay is the maximum delay between connection attempts.
	maxDelay = 5 * time.Second
)

type DBType string

const (
	MongoDBType DBType = "mongodb"
	PgSQLType   DBType = "postgresql"
)

type DBConfig struct {
	Operations DBOperations
}

type DBOperations interface {
	Connect() error
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
	mongodb.MongoOperations
	postgresql.PostgresOperations
}

func NewDBConfig(config *config.Config) (*DBConfig, error) {
	switch DBType(config.Storage.Type) {
	case MongoDBType:
		client, err := mongodb.GetMongoClient(&StorageConfig{
			Type:             config.Storage.Type,
			Host:             config.Storage.Host,
			Port:             config.Storage.Port,
			Username:         config.Storage.Username,
			Password:         config.Storage.Password,
			Database:         config.Storage.Database,
			Collections:      config.Storage.Collections,
			CollectionsUsers: config.Storage.CollectionsUsers,
		})
		if err != nil {
			return nil, err
		}
		return &DBConfig{
			&mongodb.MongoClient{
				Client: client,
				Cfg:    config,
			},
		}, nil
	case PgSQLType:
		pool, err := postgresql.NewClient(context.Background(), maxAttempts, maxDelay, &StorageConfig{
			Type:     config.Storage.Type,
			Host:     config.Storage.Host,
			Port:     config.Storage.Port,
			Username: config.Storage.Username,
			Password: config.Storage.Password,
			Database: config.Storage.Database,
		})
		if err != nil {
			return nil, err
		}
		return &DBConfig{
			&postgresql.PgClient{
				Pool: pool,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Storage.Type)
	}
}

func (s *StorageConfig) Connect() error {
	switch DBType(s.Type) {
	case MongoDBType:
		client, err := mongodb.GetMongoClient(s)
		if err != nil {
			return err
		}
		// Сохраните клиента в структуре s для дальнейшего использования.
		s.client = client
	case PgSQLType:
		pool, err := postgresql.NewClient(context.Background(), maxAttempts, maxDelay, s)
		if err != nil {
			return err
		}
		// Сохраните pool в структуре s для дальнейшего использования.
		s.pool = pool
	default:
		return fmt.Errorf("unsupported database type: %s", s.Type)
	}
	return nil
}
