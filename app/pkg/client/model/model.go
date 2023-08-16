package model

import (
	"context"
	"fmt"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/mongodb"
	"skeleton-golange-application/app/pkg/client/postgresql"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	maxConnectionAttempts = 5
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

func NewDBConfig(cfg *config.Config) (*DBConfig, error) {
	switch DBType(cfg.Storage.Type) {
	case MongoDBType:
		client, err := GetMongoClient(&StorageConfig{
			Type:             cfg.Storage.Type,
			Host:             cfg.Storage.Host,
			Port:             cfg.Storage.Port,
			Username:         cfg.Storage.Username,
			Password:         cfg.Storage.Password,
			Database:         cfg.Storage.Database,
			Collections:      cfg.Storage.Collections,
			CollectionsUsers: cfg.Storage.CollectionsUsers,
		})
		if err != nil {
			return nil, err
		}
		return &DBConfig{
			&mongodb.MongoClient{
				Client: client,
				Cfg:    cfg,
			},
		}, nil
	case PgSQLType:
		pool, err := NewClient(context.Background(), maxAttempts, maxDelay, &StorageConfig{
			Type:     cfg.Storage.Type,
			Host:     cfg.Storage.Host,
			Port:     cfg.Storage.Port,
			Username: cfg.Storage.Username,
			Password: cfg.Storage.Password,
			Database: cfg.Storage.Database,
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
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Storage.Type)
	}
}

func (s *StorageConfig) Connect() error {
	startTime := time.Now()
	metrics := NewDBPrometheusMetrics()
	metrics.DatabaseConnectionAttemptCounter.Inc()
	switch DBType(s.Type) {
	case MongoDBType:
		client, err := GetMongoClient(s)
		if err != nil {
			metrics.DatabaseConnectionFailureCounter.Inc()
			return fmt.Errorf("failed to connect to MongoDB: %w", err)
		}
		// Save the client in the s structure for future use.
		s.client = client
	case PgSQLType:
		pool, err := NewClient(context.Background(), maxAttempts, maxDelay, s)
		if err != nil {
			metrics.DatabaseConnectionFailureCounter.Inc()
			return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}
		// Save the pool in the s structure for future use.
		s.pool = pool
	default:
		return fmt.Errorf("unsupported database type: %s", s.Type)
	}
	duration := time.Since(startTime)
	metrics.ResponseTimeDBConnect.Observe(duration.Seconds())
	metrics.DatabaseConnectionSuccessCounter.Inc()
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
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Username, cfg.Password,
		cfg.Host, cfg.Port, cfg.Database,
	)
	err = postgresql.DoWithAttempts(func() error {
		ctx, cancel := context.WithTimeout(ctx, maxConnectionAttempts*time.Second)
		defer cancel()

		pgxCfg, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			log.Fatalf("Unable to parse config: %w\n", err) // <-- Use %w here.
		}

		pool, err = pgxpool.ConnectConfig(ctx, pgxCfg)
		if err != nil {
			log.Println("Failed to connect to postgres... Going to do the next attempt")
			return err
		}

		// Run database migrations
		err = postgresql.RunMigrations(dsn)
		if err != nil {
			return fmt.Errorf("failed to run migrations: %w", err) // <-- Use %w here
		}

		return nil
	}, maxAttempts, maxDelay)

	if err != nil {
		log.Fatal("All attempts are exceeded. Unable to connect to postgres")
	}

	return pool, nil
}

type DBPrometheusMetrics struct {
	DatabaseConnectionAttemptCounter prometheus.Counter
	DatabaseConnectionSuccessCounter prometheus.Counter
	DatabaseConnectionFailureCounter prometheus.Counter
	ResponseTimeDBConnect            prometheus.Histogram
}

func NewDBPrometheusMetrics() *DBPrometheusMetrics {
	return &DBPrometheusMetrics{
		DatabaseConnectionAttemptCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "database_connection_attempt_total",
			Help: "Total number of database connection attempts",
		}),
		DatabaseConnectionSuccessCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "database_connection_success_total",
			Help: "Total number of successful database connections",
		}),
		DatabaseConnectionFailureCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "database_connection_failure_total",
			Help: "Total number of failed database connections",
		}),
		ResponseTimeDBConnect: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "database_response_time_seconds",
			Help:    "Database response time histogram",
			Buckets: prometheus.DefBuckets,
		}),
		// Define additional counters for other database metrics as needed.
		// ...
	}
}
