package model

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		client, err := GetMongoClient(&StorageConfig{
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
		pool, err := NewClient(context.Background(), maxAttempts, maxDelay, &StorageConfig{
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
	DatabaseConnectionAttemptCounter.Inc()
	switch DBType(s.Type) {
	case MongoDBType:
		client, err := GetMongoClient(s)
		if err != nil {
			DatabaseConnectionFailureCounter.Inc()
			return err
		}
		// Сохраните клиента в структуре s для дальнейшего использования.
		s.client = client
	case PgSQLType:
		pool, err := NewClient(context.Background(), maxAttempts, maxDelay, s)
		if err != nil {
			DatabaseConnectionFailureCounter.Inc()
			return err
		}
		// Сохраните pool в структуре s для дальнейшего использования.
		s.pool = pool
	default:
		return fmt.Errorf("unsupported database type: %s", s.Type)
	}
	DatabaseConnectionSuccessCounter.Inc()
	return nil
}

func GetMongoClient(cfg *StorageConfig) (*mongo.Client, error) {
	connectionString := fmt.Sprintf("mongodb://%s:%s", cfg.Host, cfg.Port)
	credential := options.Credential{
		Username: cfg.Username,
		Password: cfg.Password,
	}
	clientOptions := options.Client().ApplyURI(connectionString).SetAuth(credential)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewClient(ctx context.Context, maxAttempts int, maxDelay time.Duration, cfg *StorageConfig) (pool *pgxpool.Pool, err error) {
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		cfg.Username, cfg.Password,
		cfg.Host, cfg.Port, cfg.Database,
	)

	err = postgresql.DoWithAttempts(func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		pgxCfg, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			log.Fatalf("Unable to parse config: %v\n", err)
		}

		pool, err = pgxpool.ConnectConfig(ctx, pgxCfg)
		if err != nil {
			log.Println("Failed to connect to postgres... Going to do the next attempt")
			return err
		}

		client := &postgresql.PgClient{Pool: pool}

		tables := []interface{}{config.User{}, config.Album{}}
		err = client.CheckTablePresence(tables)
		if err != nil {
			return fmt.Errorf("failed to check table presence: %v", err)
		}

		return nil
	}, maxAttempts, maxDelay)

	if err != nil {
		log.Fatal("All attempts are exceeded. Unable to connect to postgres")
	}

	return pool, nil
}

// ---------------------DB prometheus
var (
	DatabaseConnectionAttemptCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "database_connection_attempt_total",
		Help: "Total number of database connection attempts",
	})

	DatabaseConnectionSuccessCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "database_connection_success_total",
		Help: "Total number of successful database connections",
	})

	DatabaseConnectionFailureCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "database_connection_failure_total",
		Help: "Total number of failed database connections",
	})

	// Define additional counters for other database metrics as needed
	// ...

)
