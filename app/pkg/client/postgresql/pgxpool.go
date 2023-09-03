package postgresql

import (
	"context"
	"errors"
	"fmt"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/logging"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/jackc/pgx/v4"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresCollectionQuery interface {
	FindUserToEmail(email string) (config.User, error)
	CreateUser(user config.User) error
	DeleteUser(email string) error
	GetStoredRefreshToken(userEmail string) (string, error)
	SetStoredRefreshToken(userEmail, refreshToken string) error
	CreateIssue(task *config.Album) error
	CreateMany(list []config.Album) error
	GetAlbums(offset, limit int, sortBy, sortOrder, filterArtist string) ([]config.Album, int, error)
	GetIssuesByCode(code string) (config.Album, error)
	DeleteOne(code string) error
	DeleteAll() error
	UpdateIssue(album *config.Album) error
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

func (c *PgClient) CreateUser(user config.User) error {
	query := `INSERT INTO "users" (_id, name, email, password, role) VALUES ($1, $2, $3, $4, $5)`
	_, err := c.Pool.Exec(context.TODO(), query, user.ID, user.Name, user.Email, user.Password, user.Role)
	if err != nil {
		return err
	}
	return nil
}

func (c *PgClient) FindUserToEmail(email string) (config.User, error) {
	var user config.User
	query := `SELECT _id, name, email, password, role, refreshtoken FROM "users" WHERE email = $1`
	err := c.Pool.QueryRow(context.TODO(), query, email).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.RefreshToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, fmt.Errorf("user with email '%s' not found", email)
		}
		return user, err
	}
	return user, nil
}

func (c *PgClient) DeleteUser(email string) error {
	query := `DELETE FROM "users" WHERE email = $1`
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

func (c *PgClient) GetStoredRefreshToken(userEmail string) (string, error) {
	query := `SELECT refreshtoken
        FROM users
        WHERE email = $1;`
	var refreshToken string
	err := c.Pool.QueryRow(context.TODO(), query, userEmail).
		Scan(&refreshToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("user with email '%s' not found", userEmail)
		}
		return "", err
	}
	return refreshToken, nil
}

func (c *PgClient) SetStoredRefreshToken(userEmail, refreshToken string) error {
	// Update the user's refresh token
	updateQuery := `UPDATE users
        SET refreshtoken = $2
        WHERE email = $1;`
	_, err := c.Pool.Exec(context.TODO(), updateQuery, userEmail, refreshToken)
	if err != nil {
		return err
	}

	return nil
}

func (c *PgClient) CreateIssue(album *config.Album) error {
	// Check if the "album" table exists
	tableExists, err := c.TableExists("album")
	if err != nil {
		return err
	}

	if !tableExists {
		// Return an error if the table does not exist
		return fmt.Errorf("table 'album' does not exist")
	}

	query := `INSERT INTO album (_id, created_at, updated_at, title, artist, price, code, 
                   description, sender, _creator_user)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = c.Pool.Exec(context.TODO(), query, album.ID, album.CreatedAt, album.UpdatedAt, album.Title,
		album.Artist, album.Price, album.Code, album.Description, album.Sender, album.CreatorUser)
	if err != nil {
		return err
	}
	return nil
}

func (c *PgClient) CreateMany(list []config.Album) error {
	insertableList := make([]interface{}, len(list)*albumFieldCount)
	for i := range list {
		baseIndex := i * albumFieldCount
		v := &list[i] // Use a pointer to the current album.
		insertableList[baseIndex] = &v.ID
		insertableList[baseIndex+1] = &v.CreatedAt
		insertableList[baseIndex+2] = &v.UpdatedAt
		insertableList[baseIndex+3] = &v.Title
		insertableList[baseIndex+4] = &v.Artist
		insertableList[baseIndex+5] = &v.Price
		insertableList[baseIndex+6] = &v.Code
		insertableList[baseIndex+7] = &v.Description
		insertableList[baseIndex+8] = &v.Sender
		insertableList[baseIndex+9] = &v.CreatorUser
	}

	query := `INSERT INTO album (_id, created_at, updated_at, title, artist, price, code, 
                   description, sender, _creator_user) VALUES `

	var placeholders []string
	for i := 0; i < len(list); i++ {
		placeholderValues := make([]string, albumFieldCount)
		for j := 0; j < albumFieldCount; j++ {
			placeholderValues[j] = fmt.Sprintf("$%d", i*albumFieldCount+j+1)
		}
		placeholders = append(placeholders, "("+strings.Join(placeholderValues, ", ")+")")
	}

	query += strings.Join(placeholders, ", ")

	_, err := c.Pool.Exec(context.TODO(), query, insertableList...)
	if err != nil {
		return err
	}

	return nil
}

// GetAlbums Define the GetPaginatedAlbums function within your StorageOperations struct.
func (c *PgClient) GetAlbums(offset, limit int, sortBy, sortOrder, filter string) ([]config.Album, int, error) {
	// Construct the base SQL query for selecting albums
	sql := "SELECT _id, created_at, updated_at, title, artist, price, code, description, sender, _creator_user FROM album"

	// Create separate variables for WHERE and COUNT queries
	where := ""
	countTotal := "SELECT COUNT(_id) FROM album"

	// Build the WHERE clause for filtering if filter is provided
	if filter != "" {
		filterColumns := []string{"title", "artist", "code", "sender", "_creator_user"}
		quotedFilterColumns := make([]string, len(filterColumns))
		for i, col := range filterColumns {
			quotedFilterColumns[i] = fmt.Sprintf("%s ILIKE '%%%s%%'", pq.QuoteIdentifier(col), filter)
		}
		where = " WHERE " + strings.Join(quotedFilterColumns, " OR ")
		countTotal += where
	}

	// Finalize the SQL query
	sql = fmt.Sprintf("%s%s", sql, where)

	// Build the ORDER BY clause if sortOrder and sortBy are provided
	if sortOrder != "" && sortBy != "" {
		sortOrder = strings.ToUpper(sortOrder)
		if sortOrder != "ASC" && sortOrder != "DESC" {
			sortOrder = "DESC" // Default to DESC if sortOrder is invalid
		}
		sql = fmt.Sprintf("%s ORDER BY %s %s", sql, pq.QuoteIdentifier(sortBy), sortOrder)
	}

	// Add LIMIT and OFFSET to the SQL query
	sql = fmt.Sprintf("%s LIMIT %d OFFSET %d", sql, limit, offset)

	// Execute the COUNT query to get the total count
	var totalRows int
	countErr := c.Pool.QueryRow(context.TODO(), countTotal).Scan(&totalRows)
	if countErr != nil {
		return nil, 0, countErr
	}

	// Execute the main query
	rows, queryErr := c.Pool.Query(context.TODO(), sql)
	if queryErr != nil {
		return nil, 0, queryErr
	}
	defer rows.Close()

	// Process the results
	albums := make([]config.Album, 0)
	for rows.Next() {
		var album config.Album
		scanErr := rows.Scan(
			&album.ID, &album.CreatedAt, &album.UpdatedAt,
			&album.Title, &album.Artist, &album.Price,
			&album.Code, &album.Description, &album.Sender,
			&album.CreatorUser,
		)
		if scanErr != nil {
			return nil, totalRows, scanErr
		}
		albums = append(albums, album)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, totalRows, rowsErr
	}

	return albums, totalRows, nil
}

func (c *PgClient) DeleteAll() error {
	// Check if the "album" table exists.
	tableExists, err := c.TableExists("album")
	if err != nil {
		return err
	}

	if !tableExists {
		// Return nil if the table does not exist
		return nil
	}

	query := "TRUNCATE album"
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

	// Check if any row was deleted.
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no album found with code: %s", code)
	}

	return nil
}

func (c *PgClient) GetIssuesByCode(code string) (config.Album, error) {
	result := config.Album{}

	// Check if the "album" table exists.
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
		&result.Sender,
		&result.CreatorUser,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, fmt.Errorf("no album found with code: %s", code)
		}
		return result, err
	}
	return result, nil
}

func (c *PgClient) UpdateIssue(album *config.Album) error {
	// Check if the "album" table exists.
	tableExists, err := c.TableExists("album")
	if err != nil {
		return err
	}

	if !tableExists {
		// Return an error if the table does not exist
		return fmt.Errorf("table 'album' does not exist")
	}

	query := `
        UPDATE album
        SET
            created_at = $1,
            updated_at = $2,
            title = $3,
            artist = $4,
            price = $5,
            description = $6,
            sender = $7
        WHERE
            code = $8
    `
	_, err = c.Pool.Exec(context.TODO(), query,
		album.CreatedAt, time.Now(), album.Title, album.Artist, album.Price, album.Description, album.Sender, album.Code)
	if err != nil {
		return err
	}
	return nil
}
