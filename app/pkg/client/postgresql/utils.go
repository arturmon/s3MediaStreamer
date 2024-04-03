package postgresql

import (
	"context"
	"s3MediaStreamer/app/model"

	"go.opentelemetry.io/otel"

	"github.com/Masterminds/squirrel"
)

func (c *PgClient) executeSelectQuery(ctx context.Context, selectBuilder squirrel.SelectBuilder) ([]model.Track, error) {
	_, span := otel.Tracer("").Start(ctx, "executeSelectQuery")
	defer span.End()
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
				&track.Album,
				&track.AlbumArtist,
				&track.Composer,
				&track.Genre,
				&track.Lyrics,
				&track.Title,
				&track.Artist,
				&track.Year,
				&track.Comment,
				&track.Disc,
				&track.DiscTotal,
				&track.Track,
				&track.TrackTotal,
				&track.Duration,
				&track.SampleRate,
				&track.Bitrate,
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
