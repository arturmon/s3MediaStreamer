package model

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"s3MediaStreamer/app/internal/config"
	"s3MediaStreamer/app/pkg/client/postgresql"
	"s3MediaStreamer/app/pkg/logging"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type StorageConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	pool     *pgxpool.Pool
}

type DBConfig struct {
	Operations DBOperations
}

type DBOperations interface {
	Connect(_ *logging.Logger) error
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
	postgresql.PostgresOperations
}

func NewDBConfig(cfg *config.Config, logger *logging.Logger) (*DBConfig, error) {
	pool, err := NewClient(context.Background(), postgresql.MaxAttempts, postgresql.MaxDelay, &StorageConfig{
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

}

func (s *StorageConfig) Connect(logger *logging.Logger) error {
	startTime := time.Now()
	metrics := NewDBPrometheusMetrics()
	metrics.DatabaseConnectionAttemptCounter.Inc()

	pool, err := NewClient(context.Background(), postgresql.MaxAttempts, postgresql.MaxDelay, s, logger)
	if err != nil {
		metrics.DatabaseConnectionFailureCounter.Inc()
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	// Save the pool in the s structure for future use.
	s.pool = pool

	duration := time.Since(startTime)
	metrics.ResponseTimeDBConnect.Observe(duration.Seconds())
	metrics.DatabaseConnectionSuccessCounter.Inc()
	return nil
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
		ctxWithTimeout, cancel := context.WithTimeout(ctx, postgresql.MaxConnectionAttempts*time.Second)
		defer cancel()

		pgxCfg, err := pgxpool.ParseConfig(dsn.String())
		if err != nil {
			logger.Fatalf("Unable to parse config: %v\n", err)
		}
		// otel
		pgxCfg.ConnConfig.Tracer = otelpgx.NewTracer()

		pool, err = pgxpool.NewWithConfig(ctxWithTimeout, pgxCfg)
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
		logger.Println("Finish migration.")
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
