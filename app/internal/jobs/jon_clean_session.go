package jobs

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"net"
	"net/url"
)

func (j *CleanOldSessionJob) Run() {
	if !j.app.LeaderElection.Election.IsLeader() {
		j.app.Logger.Println("I'm not the leader.")
		return
	}

	j.app.Logger.Println("I'm the leader!")
	j.app.Logger.Println("init Clean old session storage...")

	var dbPool *pgxpool.Pool

	switch j.app.Cfg.Session.SessionStorageType {
	case "postgres":
		dsn := url.URL{
			Scheme:   "postgresql",
			User:     url.UserPassword(j.app.Cfg.Session.Postgresql.PostgresqlUser, j.app.Cfg.Session.Postgresql.PostgresqlPass),
			Host:     net.JoinHostPort(j.app.Cfg.Session.Postgresql.PostgresqlHost, j.app.Cfg.Session.Postgresql.PostgresqlPort),
			Path:     j.app.Cfg.Session.Postgresql.PostgresqlDatabase,
			RawQuery: "sslmode=disable", // This enables SSL/TLS
		}
		config, err := pgxpool.ParseConfig(dsn.String())
		if err != nil {
			j.app.Logger.Fatalf("Error parsing PostgreSQL config: %v", err)
		}
		dbPool, err = pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			j.app.Logger.Fatalf("Error creating PostgreSQL pool: %v", err)
		}
	default:
	}

	err := CleanSessions(dbPool)
	if err != nil {
		j.app.Logger.Fatal(err)
	}
	j.app.Logger.Printf("complete Clean old session storage.")
}

func CleanSessions(pool *pgxpool.Pool) error {
	// Build the SQL query using squirrel
	condition := squirrel.Delete("http_sessions").
		Where("expires_on < NOW()").
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := condition.ToSql()
	if err != nil {
		return err
	}

	// Execute the SQL query using the job client's DB connection
	_, err = pool.Exec(context.TODO(), query, args...)
	return err
}
