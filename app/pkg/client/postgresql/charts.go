package postgresql

import (
	"context"
	"skeleton-golange-application/model"
	"time"

	"github.com/Masterminds/squirrel"
)

// GetAlbumsForLearn retrieves all albums with 'likes' set to true.
func (c *PgClient) GetAlbumsForLearn() ([]model.Album, error) {
	// Create a new instance of squirrel.SelectBuilder
	selectBuilder := squirrel.Select("*").
		From("album").
		Where(squirrel.Eq{"likes": true}).
		Limit(ChunkSize) // Adjust the limit based on your requirements

	var albums []model.Album

	for {
		// Generate the SQL query and arguments
		sql, args, err := selectBuilder.ToSql()
		if err != nil {
			return nil, err
		}

		// Execute the SELECT query
		rows, err := c.Pool.Query(context.TODO(), sql, args...)
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

		rows.Close()

		if len(chunk) == 0 {
			// No more records to fetch
			break
		}

		albums = append(albums, chunk...)

		// Adjust the OFFSET for the next batch
		selectBuilder = selectBuilder.Offset(uint64(len(chunk)))
	}

	return albums, nil
}

// CreateTops inserts multiple records into the 'chart' table.
func (c *PgClient) CreateTops(list []model.Tops) error {
	// Create a new instance of squirrel.InsertBuilder
	insertBuilder := squirrel.Insert("chart").Columns(
		"_id",
		"created_at",
		"updated_at",
		"title",
		"artist",
		"description",
		"sender",
		"_creator_user",
	)

	for _, item := range list {
		// Use a map to specify the values for the current record
		values := map[string]interface{}{
			"_id":           item.ID,
			"created_at":    item.CreatedAt,
			"updated_at":    item.UpdatedAt,
			"title":         item.Title,
			"artist":        item.Artist,
			"description":   item.Description,
			"sender":        item.Sender,
			"_creator_user": item.CreatorUser,
		}

		// Add the current record's values to the INSERT query
		insertBuilder = insertBuilder.Values(values)
	}

	// Generate the SQL query and arguments
	sql, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	// Execute the INSERT query
	_, err = c.Pool.Exec(context.TODO(), sql, args...)
	if err != nil {
		return err
	}

	return nil
}

// CleanupRecords deletes old records from the 'chart' table based on the specified retention period.
func (c *PgClient) CleanupRecords(retentionPeriod time.Duration) error {
	// Calculate the cutoff time based on the retention period
	cutoffTime := time.Now().Add(-retentionPeriod)

	// Create a new instance of squirrel.DeleteBuilder
	deleteBuilder := squirrel.Delete("chart")

	// Add a WHERE clause to specify the condition for deleting old records
	deleteBuilder = deleteBuilder.Where(squirrel.Lt{
		"created_at": cutoffTime,
	})

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
