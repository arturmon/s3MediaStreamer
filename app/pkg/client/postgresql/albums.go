package postgresql

import (
	"context"
	"fmt"
	"skeleton-golange-application/model"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
)

// CreateAlbums inserts multiple album records into the "album" table.
func (c *PgClient) CreateAlbums(list []model.Album) error {
	if len(list) == 0 {
		return nil
	}

	// Start a transaction
	tx, err := c.Pool.Begin(context.TODO())
	if err != nil {
		return err
	}
	defer func() {
		// Defer the rollback and check for errors
		if rErr := tx.Rollback(context.TODO()); rErr != nil && err == nil {
			err = rErr
		}
	}()

	// Create a batch to batch insert queries
	batch := &pgx.Batch{}

	// Add INSERT queries to the batch for each album
	for _, album := range list {
		query := `
			INSERT INTO album (_id, created_at, updated_at, title, artist, price, code, description, sender, _creator_user, likes)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		args := []interface{}{
			album.ID,
			album.CreatedAt,
			album.UpdatedAt,
			album.Title,
			album.Artist,
			album.Price,
			album.Code,
			album.Description,
			album.Sender,
			album.CreatorUser,
			album.Likes,
		}

		batch.Queue(query, args...)
	}

	// Execute the batch
	results := c.Pool.SendBatch(context.TODO(), batch)

	// Check for errors in the batch execution
	if err = results.Close(); err != nil {
		return err
	}

	// Commit the transaction
	err = tx.Commit(context.TODO())
	if err != nil {
		return err
	}

	return nil
}

// GetAlbums retrieves a list of albums with pagination and filtering.
func (c *PgClient) GetAlbums(offset, limit int, sortBy, sortOrder, filter string) ([]model.Album, int, error) {
	// Create a new instance of squirrel.SelectBuilder
	queryBuilder := squirrel.Select("*").From("album")

	// Build the WHERE clause for filtering if filter is provided
	if filter != "" {
		filterColumns := []string{"title", "artist", "code", "sender", "_creator_user"}

		// Create a slice to hold the individual filter conditions
		var filterExprs []string
		for _, col := range filterColumns {
			// Check if exact matching is required based on the filter
			if strings.HasPrefix(filter, "=") {
				// Use "=" for exact matching
				filterExpr := fmt.Sprintf("%s = $%d", col, 1)
				filterExprs = append(filterExprs, filterExpr)
			} else {
				// Use ILIKE for pattern matching
				filterExpr := fmt.Sprintf("%s ILIKE $%d", col, 1)
				filterExprs = append(filterExprs, filterExpr)
			}
		}
		if !strings.HasPrefix(filter, "=") {
			filter = "%" + filter + "%"
		}

		// Remove the "=" from the filter value
		filter = strings.TrimPrefix(filter, "=")
		// Combine the individual filter conditions using OR
		orCondition := strings.Join(filterExprs, " OR ")

		// Then add orCondition to WHERE clause
		queryBuilder = queryBuilder.Where(orCondition, filter)
	}

	// Build the ORDER BY clause if sortBy and sortOrder are provided
	if sortBy != "" && sortOrder != "" {
		queryBuilder = queryBuilder.OrderBy(sortBy + " " + sortOrder)
	}

	// Add LIMIT and OFFSET to the query
	queryBuilder = queryBuilder.Limit(uint64(limit)).Offset(uint64(offset))

	// Generate the SQL query and arguments
	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, 0, err
	}

	// Execute the query and retrieve the results
	rows, err := c.Pool.Query(context.TODO(), sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Process the results
	var albums []model.Album
	for rows.Next() {
		var album model.Album
		err = rows.Scan(
			&album.ID, &album.CreatedAt, &album.UpdatedAt,
			&album.Title, &album.Artist, &album.Price,
			&album.Code, &album.Description, &album.Sender,
			&album.CreatorUser, &album.Likes,
		)
		if err != nil {
			return nil, 0, err
		}
		albums = append(albums, album)
	}

	// Get the total count of records (excluding pagination)
	totalRows, countErr := c.GetTotalAlbumCount(queryBuilder)
	if countErr != nil {
		return nil, 0, countErr
	}

	return albums, totalRows, nil
}

// GetAlbumsByCode retrieves an album record from the "album" table based on the provided code.
func (c *PgClient) GetAlbumsByCode(code string) (model.Album, error) {
	// Use the GetAlbums function with a filter condition to find the album by code
	albums, _, err := c.GetAlbums(0, 1, "", "", code)
	if err != nil {
		return model.Album{}, err
	}

	// Check if any album was found
	if len(albums) == 0 {
		return model.Album{}, fmt.Errorf("no album found with code: %s", code)
	}

	// Return the first album found (assuming code is unique)
	return albums[0], nil
}

// DeleteAlbums deletes a single record from the "album" table based on the provided code.
func (c *PgClient) DeleteAlbums(code string) error {
	// Create a new instance of squirrel.DeleteBuilder and specify the table name
	deleteBuilder := squirrel.Delete("album").PlaceholderFormat(squirrel.Dollar)

	// Add a WHERE condition to specify the record to delete
	deleteBuilder = deleteBuilder.Where(squirrel.Eq{"code": code})

	// Generate the SQL query and arguments
	sql, args, err := deleteBuilder.ToSql()
	if err != nil {
		return err
	}

	// Execute the DELETE query
	_, err = c.Pool.Exec(context.TODO(), sql, args...)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAlbumsAll deletes all records from the "album" table.
func (c *PgClient) DeleteAlbumsAll() error {
	// Create a new instance of squirrel.DeleteBuilder
	deleteBuilder := squirrel.Delete("").From("album")

	// Generate the SQL query and arguments
	sql, args, err := deleteBuilder.ToSql()
	if err != nil {
		return err
	}

	// Execute the DELETE query
	_, err = c.Pool.Exec(context.TODO(), sql, args...)
	if err != nil {
		return err
	}

	return nil
}

// UpdateAlbums updates an album record in the "album" table based on the provided code.
func (c *PgClient) UpdateAlbums(album *model.Album) error {
	// Create a new instance of squirrel.UpdateBuilder
	updateBuilder := squirrel.Update("album")

	// Add SET clauses to specify the columns and their new values
	updateBuilder = updateBuilder.SetMap(map[string]interface{}{
		"created_at":  album.CreatedAt,
		"updated_at":  album.UpdatedAt,
		"title":       album.Title,
		"artist":      album.Artist,
		"price":       album.Price,
		"description": album.Description,
		"sender":      album.Sender,
		"likes":       album.Likes,
	})

	// Add a WHERE condition to identify the record to update based on the provided code
	updateBuilder = updateBuilder.Where(squirrel.Eq{"code": album.Code})

	// Generate the SQL query and arguments
	sql, args, err := updateBuilder.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return err
	}

	// Execute the UPDATE query
	_, err = c.Pool.Exec(context.TODO(), sql, args...)
	if err != nil {
		return err
	}

	return nil
}
