package postgresql

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/model"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresCollectionQuery interface {
	FindUserToEmail(email string) (config.User, error)
	CreateUser(user config.User) error
	CreateIssue(task config.Album) error
	CreateMany(list []config.Album) error
	GetAllIssues() ([]config.Album, error)
	GetIssuesByCode(code string) (config.Album, error)
	DeleteOne(code string) error
	DeleteAll() error
	MarkCompleted(code string) error
}

type PostgresOperations interface {
	PostgresCollectionQuery
}

type PgClient struct {
	Pool *pgxpool.Pool
	// другие поля, если они вам нужны
}

func (c *PgClient) Begin(ctx context.Context) (pgx.Tx, error) {
	return c.Pool.Begin(ctx)
}

func (c *PgClient) FindCollections(name string) (*mongo.Collection, error) {
	return nil, fmt.Errorf("FindCollections is not supported for PostgreSQL")
}

// NewClient
func NewClient(ctx context.Context, maxAttempts int, maxDelay time.Duration, cfg *model.StorageConfig) (pool *pgxpool.Pool, err error) {
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		cfg.Username, cfg.Password,
		cfg.Host, cfg.Port, cfg.Database,
	)

	err = DoWithAttempts(func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		pgxCfg, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			log.Fatalf("Unable to parse config: %v\n", err)
		}

		// pgxCfg.ConnConfig.Logger = logrusadapter.NewLogger(logger)

		pool, err = pgxpool.ConnectConfig(ctx, pgxCfg)
		if err != nil {
			log.Println("Failed to connect to postgres... Going to do the next attempt")

			return err
		}

		return nil
	}, maxAttempts, maxDelay)

	if err != nil {
		log.Fatal("All attempts are exceeded. Unable to connect to postgres")
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

func (c *PgClient) Connect() error {
	if c.Pool != nil {
		// Acquire a connection from the pool
		conn, err := c.Pool.Acquire(context.Background())
		if err != nil {
			return err
		}
		// Release the connection back to the pool
		defer conn.Release()
		// Ping the database to check the connection
		if err := conn.Conn().Ping(context.Background()); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("pgx pool is not initialized")
	}
	return nil
}

func (c *PgClient) Ping(ctx context.Context) error {
	if c.Pool != nil {
		conn, err := c.Pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()
		err = conn.Conn().Ping(ctx)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("pgx pool is not initialized")
	}
	return nil
}

func (c *PgClient) Close(ctx context.Context) error {
	if c.Pool != nil {
		c.Pool.Close()
		c.Pool = nil
	}
	return nil
}

func (c *PgClient) CreateUser(user config.User) error {
	query := "INSERT INTO users (id, email, username, password) VALUES ($1, $2, $3, $4)"
	_, err := c.Pool.Exec(context.TODO(), query, user.Id, user.Email, user.Name, user.Password)
	if err != nil {
		return err
	}
	return nil
}

func (c *PgClient) FindUserToEmail(email string) (config.User, error) {
	var user config.User
	query := "SELECT email, username FROM users WHERE email = $1"
	err := c.Pool.QueryRow(context.TODO(), query, email).Scan(&user.Email, &user.Name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return user, fmt.Errorf("user with email '%s' not found", email)
		}
		return user, err
	}
	return user, nil
}

func (c *PgClient) CreateIssue(task config.Album) error {
	query := "INSERT INTO albums (id, name, description) VALUES ($1, $2, $3)"
	_, err := c.Pool.Exec(context.TODO(), query, task.ID, task.Artist, task.Description)
	if err != nil {
		return err
	}
	return nil
}

func (c *PgClient) CreateMany(list []config.Album) error {
	insertableList := make([]interface{}, len(list))
	for i, v := range list {
		insertableList[i] = []interface{}{v.ID, v.Artist, v.Description}
	}
	query := "INSERT INTO albums (id, name, description) VALUES "
	var values []interface{}
	for i := range insertableList {
		values = append(values, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
	}
	query += fmt.Sprintf("%s", values)
	_, err := c.Pool.Exec(context.TODO(), query, insertableList...)
	if err != nil {
		return err
	}
	return nil
}

func (c *PgClient) GetAllIssues() ([]config.Album, error) {
	query := "SELECT code, title, description FROM albums"
	rows, err := c.Pool.Query(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []config.Album
	for rows.Next() {
		var issue config.Album
		err := rows.Scan(&issue.Code, &issue.Title, &issue.Description)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return issues, nil
}

func (c *PgClient) DeleteAll() error {
	query := "DELETE FROM albums"
	_, err := c.Pool.Exec(context.TODO(), query)
	if err != nil {
		return err
	}

	return nil
}

func (c *PgClient) DeleteOne(code string) error {
	query := "DELETE FROM albums WHERE code = $1"
	_, err := c.Pool.Exec(context.TODO(), query, code)
	if err != nil {
		return err
	}
	return nil
}

func (c *PgClient) MarkCompleted(code string) error {
	query := "UPDATE albums SET completed = true WHERE code = $1"
	_, err := c.Pool.Exec(context.TODO(), query, code)
	if err != nil {
		return err
	}

	return nil
}

func (c *PgClient) GetIssuesByCode(code string) (config.Album, error) {
	result := config.Album{}
	query := "SELECT * FROM albums WHERE code = $1"
	row := c.Pool.QueryRow(context.TODO(), query, code)
	err := row.Scan(
		&result.ID,
		&result.Code,
		&result.Artist,
		&result.Description,
		&result.Completed,
	)
	if err != nil {
		if pgx.ErrNoRows == err {
			return result, fmt.Errorf("no album found with code: %s", code)
		}
		return result, err
	}
	return result, nil
}
