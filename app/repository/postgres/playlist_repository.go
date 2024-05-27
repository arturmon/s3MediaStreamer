package postgres

import (
	"context"
	"errors"
	"s3MediaStreamer/app/model"

	"go.opentelemetry.io/otel"

	"github.com/Masterminds/squirrel"
)

type PlaylistRepositoryInterface interface {
	CreatePlayListName(ctx context.Context, newPlaylist model.PLayList) error
	GetPlayListByID(ctx context.Context, playlistID string) (model.PLayList, []model.Track, error)
	DeletePlaylist(ctx context.Context, playlistID string) error
	PlaylistExists(ctx context.Context, playlistID string) bool
	ClearPlayList(ctx context.Context, playlistID string) error
	UpdatePlaylistTrackOrder(ctx context.Context, playlistID string, trackOrderRequest []string) error
	GetTracksByPlaylist(ctx context.Context, playlistID string) ([]model.Track, error)
	GetAllPlayList(ctx context.Context, creatorUserID string) ([]model.PLayList, error)
	GetUserAtPlayList(ctx context.Context, playlistID string) (string, error)
}

func (c *Client) CreatePlayListName(ctx context.Context, playlist model.PLayList) error {
	_, span := otel.Tracer("").Start(ctx, "CreatePlayListName")
	defer span.End()
	// Start a transaction
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// Defer the rollback and check for errors
		if rErr := tx.Rollback(ctx); rErr != nil && err == nil {
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
	_, err = c.Pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	// Commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetPlayListByID(ctx context.Context, playlistID string) (model.PLayList, []model.Track, error) {
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
	err = c.Pool.QueryRow(ctx, sql, args...).
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

func (c *Client) DeletePlaylist(ctx context.Context, playlistID string) error {
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
	_, err = c.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) PlaylistExists(ctx context.Context, playlistID string) bool {
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
	err := c.Pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return false // An error occurred or playlist does not exist
	}
	// If count > 0, the playlist exists
	if count > 0 {
		return true
	}

	return false
}

func (c *Client) ClearPlayList(ctx context.Context, playlistID string) error {
	_, span := otel.Tracer("").Start(ctx, "ClearPlayList")
	defer span.End()
	// Start a transaction
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// Defer the rollback and check for errors
		if rErr := tx.Rollback(ctx); rErr != nil && err == nil {
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
	_, err = c.Pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	// Commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

// UpdatePlaylistTrackOrder updates the order of tracks within a playlist based on the provided order.
func (c *Client) UpdatePlaylistTrackOrder(ctx context.Context, playlistID string, trackOrderRequest []string) error {
	_, span := otel.Tracer("").Start(ctx, "UpdatePlaylistTrackOrder")
	defer span.End()
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && err == nil {
			err = rErr
		}
	}()

	// Fetch existing tracks in the playlist and their positions
	existingTracks := make(map[string]int)
	rows, err := tx.Query(ctx, "SELECT reference_id, position FROM playlist_tracks WHERE playlist_id = $1", playlistID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var trackID string
		var position int
		if err = rows.Scan(&trackID, &position); err != nil {
			return err
		}
		existingTracks[trackID] = position
	}

	// Determine the last position in the playlist
	lastPosition := 0
	for _, pos := range existingTracks {
		if pos > lastPosition {
			lastPosition = pos
		}
	}

	// Start updating the track positions based on the provided order
	for _, trackID := range trackOrderRequest {
		// Check if the track is already in the playlist
		if _, exists := existingTracks[trackID]; !exists {
			// Track is not in the playlist, so insert it with the next position
			lastPosition++
			_, err = tx.Exec(ctx, "INSERT INTO playlist_tracks (playlist_id, reference_type, reference_id, position) VALUES ($1, $2, $3, $4)",
				playlistID, "track", trackID, lastPosition)
			if err != nil {
				return err
			}
		}
	}

	// Commit the transaction.
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

// GetTracksByPlaylist retrieves tracks associated with a playlist.
func (c *Client) GetTracksByPlaylist(ctx context.Context, playlistID string) ([]model.Track, error) {
	_, span := otel.Tracer("").Start(ctx, "GetTracksByPlaylist")
	defer span.End()
	// Create a SQL query to fetch tracks associated with the given playlist
	selectQuery := squirrel.Select("t.*").
		From("tracks t").
		Join("playlist_tracks pt ON t._id = pt.reference_id").                                                     // Changed "pt.track_id" to "pt.reference_id"
		Where(squirrel.And{squirrel.Eq{"pt.playlist_id": playlistID}, squirrel.Eq{"pt.reference_type": "track"}}). // Added condition for reference_type
		PlaceholderFormat(squirrel.Dollar)

	// Convert the SQL query to SQL and arguments
	sql, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, err
	}

	// Execute the query and scan the result into a slice of tracks
	rows, err := c.Pool.Query(ctx, sql, args...)
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
		)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}
	return tracks, nil
}

// GetAllPlayList retrieves tracks associated with a playlist.
func (c *Client) GetAllPlayList(ctx context.Context, creatorUserID string) ([]model.PLayList, error) {
	_, span := otel.Tracer("").Start(ctx, "GetAllPlayList")
	defer span.End()
	// Initialize an empty playlists to store the result
	var playlists []model.PLayList

	// Create a SQL query to fetch the playlist by its ID
	selectQuery := squirrel.Select("*").From("playlists").
		PlaceholderFormat(squirrel.Dollar)

	// directive
	if creatorUserID != "admin" {
		selectQuery = selectQuery.Where(squirrel.Eq{"_creator_user": creatorUserID})
	}

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
		// if playlist.CreatorUser == con_user_id {
		playlists = append(playlists, playlist)
		//}
	}

	// Check for any error that were encountered during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return playlists, nil
}

func (c *Client) GetUserAtPlayList(ctx context.Context, playlistID string) (string, error) {
	_, span := otel.Tracer("").Start(ctx, "GetUserAtPlayList")
	defer span.End()

	// Create a SQL query to fetch the playlist by its ID
	selectQuery := squirrel.Select("_creator_user").From("playlists").
		Where(squirrel.Eq{"_id": playlistID}).
		PlaceholderFormat(squirrel.Dollar)

	// Convert the SQL query to SQL and arguments
	sql, args, err := selectQuery.ToSql()
	if err != nil {
		return "", err
	}

	// Execute the query and scan the result into the playlist struct
	rows, err := c.Pool.Query(ctx, sql, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if err != nil {
		return "", err
	}

	var creatorUser string
	if !rows.Next() {
		return "", errors.New("playlist not found")
	}
	// Scan the _creator_user value from the row into the creatorUser variable
	if err = rows.Scan(&creatorUser); err != nil {
		return "", err
	}

	return creatorUser, nil
}
