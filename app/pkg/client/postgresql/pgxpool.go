package postgresql

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"skeleton-golange-application/app/internal/config"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresCollectionQuery interface {
	FindUserToEmail(email string) (config.User, error)
	CreateUser(user config.User) error
	DeleteUser(email string) error
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
	return nil, fmt.Errorf("FindCollections is not supported for PostgreSQL")
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
	query := `INSERT INTO "user" (_id, name, email, password) VALUES ($1, $2, $3, $4)`
	_, err := c.Pool.Exec(context.TODO(), query, user.Id, user.Name, user.Email, user.Password)
	if err != nil {
		return err
	}
	return nil
}

func (c *PgClient) FindUserToEmail(email string) (config.User, error) {
	var user config.User
	query := `SELECT _id, name, email, password FROM "user" WHERE email = $1`
	err := c.Pool.QueryRow(context.TODO(), query, email).Scan(&user.Id, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if err == pgx.ErrNoRows {
			return user, fmt.Errorf("user with email '%s' not found", email)
		}
		return user, err
	}
	return user, nil
}

func (c *PgClient) DeleteUser(email string) error {
	query := `DELETE FROM "user" WHERE email = $1`
	result, err := c.Pool.Exec(context.TODO(), query, email)
	if err != nil {
		return err
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user with email '%s' not found", email)
	}
	return nil
}

func (c *PgClient) CreateIssue(album config.Album) error {
	// Check if the "album" table exists
	tableExists, err := c.TableExists("album")
	if err != nil {
		return err
	}

	if !tableExists {
		// Return an error if the table does not exist
		return fmt.Errorf("table 'album' does not exist")
	}

	query := `INSERT INTO album (_id, created_at, updated_at, title, artist, price, code, description, completed)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err = c.Pool.Exec(context.TODO(), query, album.ID, album.CreatedAt, album.UpdatedAt, album.Title,
		album.Artist, album.Price, album.Code, album.Description, album.Completed)
	if err != nil {
		return err
	}
	return nil
}

func (c *PgClient) CreateMany(list []config.Album) error {
	insertableList := make([]interface{}, 0)
	for _, v := range list {
		insertableList = append(insertableList, v.ID, v.CreatedAt, v.UpdatedAt, v.Title, v.Artist, v.Price, v.Code, v.Description, v.Completed)
	}

	query := `INSERT INTO album (_id, created_at, updated_at, title, artist, price, code, description, completed) VALUES `

	var placeholders []string
	for i := 0; i < len(list); i++ {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, i*9+6, i*9+7, i*9+8, i*9+9))
	}

	query += strings.Join(placeholders, ", ")

	_, err := c.Pool.Exec(context.TODO(), query, insertableList...)
	if err != nil {
		return err
	}

	return nil
}

func (c *PgClient) GetAllIssues() ([]config.Album, error) {
	tableExists, err := c.TableExists("album")
	if err != nil {
		return nil, err
	}

	if !tableExists {
		// Return an empty slice if the table does not exist
		return make([]config.Album, 0), nil
	}

	query := "SELECT _id, created_at, updated_at, title, artist, price, code, description, completed FROM album"
	rows, err := c.Pool.Query(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	albums := make([]config.Album, 0) // Initialize an empty slice
	for rows.Next() {
		var album config.Album
		err := rows.Scan(&album.ID, &album.CreatedAt, &album.UpdatedAt, &album.Title, &album.Artist, &album.Price, &album.Code, &album.Description, &album.Completed)
		if err != nil {
			return nil, err
		}
		albums = append(albums, album)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return albums, nil
}

func (c *PgClient) DeleteAll() error {
	// Check if the "album" table exists
	tableExists, err := c.TableExists("album")
	if err != nil {
		return err
	}

	if !tableExists {
		// Return nil if the table does not exist
		return nil
	}

	query := "DELETE FROM album"
	_, err = c.Pool.Exec(context.TODO(), query)
	if err != nil {
		return err
	}

	return nil
}

func (c *PgClient) DeleteOne(code string) error {
	query := "DELETE FROM album WHERE code = $1"
	commandTag, err := c.Pool.Exec(context.TODO(), query, code)
	if err != nil {
		return err
	}

	// Check if any row was deleted
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no album found with code: %s", code)
	}

	return nil
}

func (c *PgClient) MarkCompleted(code string) error {
	query := "UPDATE album SET completed = true WHERE code = $1"
	_, err := c.Pool.Exec(context.TODO(), query, code)
	if err != nil {
		return err
	}

	return nil
}

func (c *PgClient) GetIssuesByCode(code string) (config.Album, error) {
	result := config.Album{}

	// Check if the "album" table exists
	tableExists, err := c.TableExists("album")
	if err != nil {
		return result, err
	}

	if !tableExists {
		// Return an empty result if the table does not exist
		return result, nil
	}

	query := "SELECT * FROM album WHERE code = $1"
	row := c.Pool.QueryRow(context.TODO(), query, code)
	err = row.Scan(
		&result.ID,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.Title,
		&result.Artist,
		&result.Price,
		&result.Code,
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
