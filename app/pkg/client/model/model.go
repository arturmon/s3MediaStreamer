package model

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/mongodb"
	"skeleton-golange-application/app/pkg/client/postgresql"
	"skeleton-golange-application/app/pkg/logging"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
	Connect(_ *logging.Logger) error
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
	mongodb.MongoOperations
	postgresql.PostgresOperations
}

func NewDBConfig(cfg *config.Config, logger *logging.Logger) (*DBConfig, error) {
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
		}, logger)
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
		}, logger)
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

func (s *StorageConfig) Connect(logger *logging.Logger) error {
	startTime := time.Now()
	metrics := NewDBPrometheusMetrics()
	metrics.DatabaseConnectionAttemptCounter.Inc()
	switch DBType(s.Type) {
	case MongoDBType:
		client, err := GetMongoClient(s, logger)
		if err != nil {
			metrics.DatabaseConnectionFailureCounter.Inc()
			return fmt.Errorf("failed to connect to MongoDB: %w", err)
		}
		// Save the client in the s structure for future use.
		s.client = client
	case PgSQLType:
		pool, err := NewClient(context.Background(), maxAttempts, maxDelay, s, logger)
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

func GetMongoClient(cfg *StorageConfig, logger *logging.Logger) (*mongo.Client, error) {
	connectionString := net.JoinHostPort(cfg.Host, cfg.Port)
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
	logger.Println("Connect MongoDB")
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewClient(ctx context.Context, maxAttempts int,
	maxDelay time.Duration, cfg *StorageConfig, logger *logging.Logger) (*pgxpool.Pool, error) {
	dsn := url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(cfg.Username, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, cfg.Port),
		Path:     cfg.Database,
		RawQuery: "sslmode=disable", // This enables SSL/TLS
	}

	var pool *pgxpool.Pool

	err := postgresql.DoWithAttempts(func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, maxConnectionAttempts*time.Second)
		defer cancel()

		pgxCfg, err := pgxpool.ParseConfig(dsn.String())
		if err != nil {
			logger.Fatalf("Unable to parse config: %v\n", err)
		}

		pool, err = pgxpool.ConnectConfig(ctxWithTimeout, pgxCfg)
		if err != nil {
			logger.Println("Failed to connect to postgres... Going to do the next attempt")
			return err
		}

		// Run database migrations
		logger.Println("Start migration...")
		err = postgresql.RunMigrations(ctx, dsn.String())
		if err != nil {
			logger.Printf("Error: %s", err)
			return err
		}

		return nil
	}, maxAttempts, maxDelay)

	if err != nil {
		logger.Fatal("All attempts are exceeded. Unable to connect to postgres")
		return nil, err
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
