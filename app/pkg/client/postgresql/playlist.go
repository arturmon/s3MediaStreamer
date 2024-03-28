package postgresql

import (
	"context"
	"s3MediaStreamer/app/model"

	"go.opentelemetry.io/otel"

	"github.com/Masterminds/squirrel"
)

func (c *PgClient) CreatePlayListName(ctx context.Context, playlist model.PLayList) error {
	_, span := otel.Tracer("").Start(ctx, "CreatePlayListName")
	defer span.End()
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

	// Create a new squirrel.InsertBuilder
	insertBuilder := squirrel.
		Insert("playlists").
		Columns("_id", "created_at", "level", "title", "description", "_creator_user").
		Values(
			playlist.ID,
			playlist.CreatedAt,
			playlist.Level,
			playlist.Title,
			playlist.Description,
			playlist.CreatorUser,
		)
	insertBuilder = insertBuilder.PlaceholderFormat(squirrel.Dollar)
	// Get the SQL query and args from the InsertBuilder
	query, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	// Execute the query
	_, err = c.Pool.Exec(context.TODO(), query, args...)
	if err != nil {
		return err
	}

	// Commit the transaction
	err = tx.Commit(context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (c *PgClient) GetPlayListByID(ctx context.Context, playlistID string) (model.PLayList, []model.Track, error) {
	_, span := otel.Tracer("").Start(ctx, "GetPlayListByID")
	defer span.End()
	// Initialize an empty playlist to store the result
	var playlist model.PLayList
	var tracks []model.Track

	// Create a SQL query to fetch the playlist by its ID
	selectQuery := squirrel.Select("*").From("playlists").
		Where(squirrel.Eq{"_id": playlistID}).
		PlaceholderFormat(squirrel.Dollar)

	// Convert the SQL query to SQL and arguments
	sql, args, err := selectQuery.ToSql()
	if err != nil {
		return playlist, tracks, err
	}

	// Execute the query and scan the result into the playlist struct
	err = c.Pool.QueryRow(context.TODO(), sql, args...).
		Scan(&playlist.ID, &playlist.CreatedAt, &playlist.Level, &playlist.Title, &playlist.Description, &playlist.CreatorUser)

	if err != nil {
		return playlist, tracks, err
	}

	// Retrieve the tracks associated with the playlist and add them to the playlist
	tracks, err = c.GetTracksByPlaylist(ctx, playlistID)
	if err != nil {
		return playlist, tracks, err
	}

	return playlist, tracks, nil
}

func (c *PgClient) DeletePlaylist(ctx context.Context, playlistID string) error {
	_, span := otel.Tracer("").Start(ctx, "DeletePlaylist")
	defer span.End()
	deleteBuilder := squirrel.Delete("playlists").PlaceholderFormat(squirrel.Dollar)

	// Add a WHERE condition to specify the record to delete
	deleteBuilder = deleteBuilder.Where(squirrel.Eq{"_id": playlistID})
	deleteBuilder = deleteBuilder.PlaceholderFormat(squirrel.Dollar)

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

func (c *PgClient) PlaylistExists(ctx context.Context, playlistID string) bool {
	_, span := otel.Tracer("").Start(ctx, "PlaylistExists")
	defer span.End()
	// Create a SELECT query using Squirrel to count rows in the playlists table
	queryBuilder := squirrel.
		Select("COUNT(*)").
		From("playlists").
		Where(squirrel.Eq{"_id": playlistID})

	// Generate the SQL query and arguments
	query, args, _ := queryBuilder.PlaceholderFormat(squirrel.Dollar).ToSql()

	// Execute the query to count rows
	var count int
	err := c.Pool.QueryRow(context.TODO(), query, args...).Scan(&count)
	if err != nil {
		return false // An error occurred or playlist does not exist
	}
	// If count > 0, the playlist exists
	if count > 0 {
		return true
	}

	return false
}

func (c *PgClient) ClearPlayList(ctx context.Context, playlistID string) error {
	_, span := otel.Tracer("").Start(ctx, "ClearPlayList")
	defer span.End()
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

	// Create a DELETE query using Squirrel to remove all tracks from the playlist
	deleteBuilder := squirrel.
		Delete("playlist_tracks").
		Where(squirrel.Eq{"playlist_id": playlistID})

	// Generate the SQL query and arguments
	query, args, err := deleteBuilder.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return err
	}

	// Execute the DELETE query
	_, err = c.Pool.Exec(context.TODO(), query, args...)
	if err != nil {
		return err
	}

	// Commit the transaction
	err = tx.Commit(context.TODO())
	if err != nil {
		return err
	}

	return nil
}

// UpdatePlaylistTrackOrder updates the order of tracks within a playlist based on the provided order.
func (c *PgClient) UpdatePlaylistTrackOrder(ctx context.Context, playlistID string, trackOrderRequest []string) error {
	_, span := otel.Tracer("").Start(ctx, "UpdatePlaylistTrackOrder")
	defer span.End()
	tx, err := c.Pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(context.Background()); rErr != nil && err == nil {
			err = rErr
		}
	}()

	// Create a map to store the new positions for each track
	newPositions := make(map[string]int)
	positionCounter := 1

	// Assign positions to tracks based on the provided order
	for _, trackID := range trackOrderRequest {
		newPositions[trackID] = positionCounter
		positionCounter++
	}

	// Start updating the track positions based on the provided order
	for _, trackID := range trackOrderRequest {
		// Check if the track exists in the new order
		newPosition, exists := newPositions[trackID]
		if exists {
			// Use the ON CONFLICT clause to handle conflicts by inserting the track
			updateQuery := squirrel.Insert("playlist_tracks").
				Columns("playlist_id", "track_id", "position").
				Values(playlistID, trackID, newPosition).
				Suffix("ON CONFLICT (playlist_id, track_id) DO UPDATE SET position = EXCLUDED.position")

			sql, args, errInsertQuery := updateQuery.PlaceholderFormat(squirrel.Dollar).ToSql()
			if errInsertQuery != nil {
				return err
			}

			_, errQuery := c.Pool.Exec(context.TODO(), sql, args...)
			if errQuery != nil {
				return errQuery
			}
		}
	}

	// Commit the transaction.
	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// GetTracksByPlaylist retrieves tracks associated with a playlist.
func (c *PgClient) GetTracksByPlaylist(ctx context.Context, playlistID string) ([]model.Track, error) {
	_, span := otel.Tracer("").Start(ctx, "GetTracksByPlaylist")
	defer span.End()
	// Create a SQL query to fetch tracks associated with the given playlist
	selectQuery := squirrel.Select("t.*").
		From("tracks t").
		Join("playlist_tracks pt ON t._id = pt.track_id").
		Where(squirrel.Eq{"pt.playlist_id": playlistID}).
		PlaceholderFormat(squirrel.Dollar)

	// Convert the SQL query to SQL and arguments
	sql, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, err
	}

	// Execute the query and scan the result into a slice of tracks
	rows, err := c.Pool.Query(context.TODO(), sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tracks []model.Track
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
			&track.Sender,
			&track.CreatorUser,
			&track.Likes,
			&track.S3Version,
		)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}
	return tracks, nil
}

// GetAllPlayList retrieves tracks associated with a playlist.
func (c *PgClient) GetAllPlayList(ctx context.Context) ([]model.PLayList, error) {
	_, span := otel.Tracer("").Start(ctx, "GetAllPlayList")
	defer span.End()
	// Initialize an empty playlists to store the result
	var playlists []model.PLayList

	// Create a SQL query to fetch the playlist by its ID
	selectQuery := squirrel.Select("*").From("playlists").
		PlaceholderFormat(squirrel.Dollar)

	// Convert the SQL query to SQL and arguments
	sql, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, err
	}

	// Execute the query
	rows, err := c.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the result set
	for rows.Next() {
		var playlist model.PLayList
		err = rows.Scan(&playlist.ID, &playlist.CreatedAt, &playlist.Level, &playlist.Title, &playlist.Description, &playlist.CreatorUser)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, playlist)
	}

	// Check for any error that were encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return playlists, nil
}
