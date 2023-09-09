package postgresql

import (
	"context"
	"errors"
	"fmt"
	"skeleton-golange-application/app/pkg/logging"
	"skeleton-golange-application/model"
	"strings"
	"time"

	"github.com/fatih/structs"

	"github.com/lib/pq"

	"github.com/jackc/pgx/v4"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresCollectionQuery interface {
	FindUserByType(value interface{}, columnType string) (model.User, error)
	CreateUser(user model.User) error
	DeleteUser(email string) error
	GetStoredRefreshToken(userEmail string) (string, error)
	SetStoredRefreshToken(userEmail, refreshToken string) error
	UpdateUserFieldsByEmail(email string, fields map[string]interface{}) error
	CreateIssue(task *model.Album) error
	CreateMany(list []model.Album) error
	GetAlbums(offset, limit int, sortBy, sortOrder, filterArtist string) ([]model.Album, int, error)
	GetIssuesByCode(code string) (model.Album, error)
	DeleteOne(code string) error
	DeleteAll() error
	UpdateIssue(album *model.Album) error
	GetAllAlbumsForLearn() ([]model.Album, error)
	CreateManyTops(list []model.Tops) error
	CleanupOldRecords(retentionPeriod time.Duration) error
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

func (c *PgClient) CreateUser(user model.User) error {
	query := `INSERT INTO "users" (_id, name, email, password, role) VALUES ($1, $2, $3, $4, $5)`
	_, err := c.Pool.Exec(context.TODO(), query, user.ID, user.Name, user.Email, user.Password, user.Role)
	if err != nil {
		return err
	}
	return nil
}

func (c *PgClient) FindUserByType(value interface{}, columnType string) (model.User, error) {
	var user model.User
	query := fmt.Sprintf(`SELECT _id, name, email, password, role, refreshtoken, Otp_enabled, Otp_secret, Otp_auth_url FROM "users" WHERE %s = $1`, columnType)
	err := c.Pool.QueryRow(context.TODO(), query, value).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role,
			&user.RefreshToken, &user.Otp_enabled, &user.Otp_secret, &user.Otp_auth_url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, fmt.Errorf("user not found")
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

func (c *PgClient) CreateIssue(album *model.Album) error {
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
                   description, sender, _creator_user, likes)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err = c.Pool.Exec(context.TODO(), query, album.ID, album.CreatedAt, album.UpdatedAt, album.Title,
		album.Artist, album.Price, album.Code, album.Description, album.Sender, album.CreatorUser, album.Likes)
	if err != nil {
		return err
	}
	return nil
}

func (c *PgClient) CreateMany(list []model.Album) error {
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
func (c *PgClient) GetAlbums(offset, limit int, sortBy, sortOrder, filter string) ([]model.Album, int, error) {
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
	albums := make([]model.Album, 0)
	for rows.Next() {
		var album model.Album
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

func (c *PgClient) GetIssuesByCode(code string) (model.Album, error) {
	result := model.Album{}

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
		&result.Likes,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, fmt.Errorf("no album found with code: %s", code)
		}
		return result, err
	}
	return result, nil
}

func (c *PgClient) UpdateIssue(album *model.Album) error {
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
            sender = $7,
        	likes = $8
        WHERE
            code = $9
    `
	_, err = c.Pool.Exec(context.TODO(), query,
		album.CreatedAt, time.Now(), album.Title, album.Artist, album.Price, album.Description, album.Sender, album.Likes, album.Code)
	if err != nil {
		return err
	}
	return nil
}

func (c *PgClient) UpdateUserFieldsByEmail(email string, fields map[string]interface{}) error {
	// Construct the SET clause for the SQL query dynamically.
	setClause := "SET "
	values := make([]interface{}, 0, len(fields)+1) // +1 for the email parameter

	i := 1
	for key, value := range fields {
		setClause += key + " = $" + fmt.Sprint(i) + ", "
		values = append(values, value)
		i++
	}

	// Remove the trailing comma and space.
	setClause = setClause[:len(setClause)-2]

	// Construct the SQL query.
	query := `
        UPDATE users
        ` + setClause + `
        WHERE email = $` + fmt.Sprint(i) + `
    `

	// Add the email parameter value to the values slice.
	values = append(values, email)

	// Execute the SQL query.
	_, err := c.Pool.Exec(context.TODO(), query, values...)
	if err != nil {
		return err
	}

	return nil
}

func (c *PgClient) GetAllAlbumsForLearn() ([]model.Album, error) {
	query := "SELECT * FROM album WHERE likes = true LIMIT $1 OFFSET $2" // Updated query
	offset := 0
	var albums []model.Album

	for {
		rows, err := c.Pool.Query(context.TODO(), query, ChunkSize, offset)
		if err != nil {
			return nil, err
		}
		var chunk []model.Album
		for rows.Next() {
			var album model.Album
			err = rows.Scan(
				&album.ID,
				&album.CreatedAt,
				&album.UpdatedAt,
				&album.Title,
				&album.Artist,
				&album.Price,
				&album.Code,
				&album.Description,
				&album.Sender,
				&album.CreatorUser,
				&album.Likes,
			)
			if err != nil {
				return nil, err
			}
			chunk = append(chunk, album)
		}
		albums = append(albums, chunk...)
		rows.Close()
		if len(chunk) < ChunkSize {
			break
		}
		offset += ChunkSize
	}
	return albums, nil
}

func (c *PgClient) CreateManyTops(list []model.Tops) error {
	if len(list) == 0 {
		return nil
	}

	query := `INSERT INTO chart (_id, created_at, updated_at, title, artist, description, sender, _creator_user) VALUES `

	var placeholders []string
	var insertableValues []interface{}

	for i, item := range list {
		itemFields := structs.Fields(item)

		for _, field := range itemFields {
			insertableValues = append(insertableValues, field.Value())
		}

		placeholderValues := make([]string, len(itemFields)) // Use len() to get the length of itemFields
		for j := 0; j < len(itemFields); j++ {
			placeholderValues[j] = fmt.Sprintf("$%d", (i*len(itemFields))+j+1)
		}
		placeholders = append(placeholders, "("+strings.Join(placeholderValues, ", ")+")")
	}

	query += strings.Join(placeholders, ", ")

	_, err := c.Pool.Exec(context.TODO(), query, insertableValues...)
	if err != nil {
		return err
	}

	return nil
}

func (c *PgClient) CleanupOldRecords(retentionPeriod time.Duration) error {
	cutoffTime := time.Now().Add(-retentionPeriod)
	query := `
        DELETE FROM chart
        WHERE created_at < $1
    `

	_, err := c.Pool.Exec(context.TODO(), query, cutoffTime)
	if err != nil {
		return err
	}

	return nil
}
