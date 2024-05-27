package connect

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/repository/postgres"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	MaxConnectionAttempts = 5
	// MaxAttempts is the maximum number of attempts to connect to the database.
	MaxAttempts = 10
	// MaxDelay is the maximum delay between connection attempts.
	MaxDelay = 5 * time.Second
)

type StorageConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	pool     *pgxpool.Pool
}

func NewDBConfig(ctx context.Context, cfg *model.Config, logger *logs.Logger) (*postgres.Client, error) {
	logger.Info("Starting Postgres Connection...")
	pool, err := NewClient(ctx, MaxAttempts, MaxDelay, &StorageConfig{
		Host:     cfg.Storage.Host,
		Port:     cfg.Storage.Port,
		Username: cfg.Storage.Username,
		Password: cfg.Storage.Password,
		Database: cfg.Storage.Database,
	}, logger)
	if err != nil {
		return nil, err
	}
	return &postgres.Client{
		Pool: pool,
	}, nil
}

func (s *StorageConfig) Connect(ctx context.Context, logger *logs.Logger) error {
	startTime := time.Now()
	metrics := NewDBPrometheusMetrics()
	metrics.DatabaseConnectionAttemptCounter.Inc()

	pool, err := NewClient(ctx, MaxAttempts, MaxDelay, s, logger)
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
	maxDelay time.Duration, cfg *StorageConfig, logger *logs.Logger) (*pgxpool.Pool, error) {
	dsn := url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(cfg.Username, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, cfg.Port),
		Path:     cfg.Database,
		RawQuery: "sslmode=disable", // This enables SSL/TLS
	}

	var pool *pgxpool.Pool

	err := DoWithAttempts(func() error {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, MaxConnectionAttempts*time.Second)
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
		err = RunMigrations(ctx, dsn.String())
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
