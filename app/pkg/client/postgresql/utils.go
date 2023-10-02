package postgresql

import (
	"context"
	"skeleton-golange-application/model"

	"github.com/Masterminds/squirrel"
)

func (c *PgClient) executeSelectQuery(selectBuilder squirrel.SelectBuilder) ([]model.Album, error) {
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
