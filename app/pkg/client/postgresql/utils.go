package postgresql

import (
	"context"
	"skeleton-golange-application/app/model"

	"github.com/Masterminds/squirrel"
)

func (c *PgClient) executeSelectQuery(selectBuilder squirrel.SelectBuilder) ([]model.Track, error) {
	var tracks []model.Track

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

		var chunk []model.Track

		for rows.Next() {
			var track model.Track
			err = rows.Scan(
				&track.ID,
				&track.CreatedAt,
				&track.UpdatedAt,
				&track.Title,
				&track.Artist,
				&track.Description,
				&track.Sender,
				&track.CreatorUser,
				&track.Likes,
				&track.S3Version,
			)
			if err != nil {
				return nil, err
			}
			chunk = append(chunk, track)
		}

		rows.Close()

		if len(chunk) == 0 {
			// No more records to fetch
			break
		}

		tracks = append(tracks, chunk...)

		// Adjust the OFFSET for the next batch
		selectBuilder = selectBuilder.Offset(uint64(len(chunk)))
	}

	return tracks, nil
}
