package postgresql

import (
	"context"
	"fmt"
	"skeleton-golange-application/app/model"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

// CreateTracks inserts multiple track records into the "track" table.
func (c *PgClient) CreateTracks(list []model.Track) error {
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

	// Add INSERT queries to the batch for each track
	for _, track := range list {
		query := `
			INSERT INTO tracks (_id, created_at, updated_at, album, album_artist, composer, genre, lyrics, title, artist, year, comment, disc, disc_total, track, track_total, duration, sample_rate, bitrate, sender, _creator_user, likes, s3Version)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
		`
		args := []interface{}{
			track.ID,
			track.CreatedAt,
			track.UpdatedAt,
			track.Album,
			track.AlbumArtist,
			track.Composer,
			track.Genre,
			track.Lyrics,
			track.Title,
			track.Artist,
			track.Year,
			track.Comment,
			track.Disc,
			track.DiscTotal,
			track.Track,
			track.TrackTotal,
			track.Duration,
			track.SampleRate,
			track.Bitrate,
			track.Sender,
			track.CreatorUser,
			track.Likes,
			track.S3Version,
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

// GetTracks retrieves a list of tracks with pagination and filtering.
func (c *PgClient) GetTracks(offset, limit int, sortBy, sortOrder, filter string) ([]model.Track, int, error) {
	// Create a new instance of squirrel.SelectBuilder
	queryBuilder := squirrel.Select("*").From("tracks")

	// Build the WHERE clause for filtering if filter is provided
	if filter != "" {
		filterColumns := []string{"title", "artist", "sender", "_creator_user"}

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
	var tracks []model.Track
	for rows.Next() {
		var track model.Track
		err = rows.Scan(
			&track.ID, &track.CreatedAt, &track.UpdatedAt,
			&track.Album, &track.AlbumArtist, &track.Composer,
			&track.Genre, &track.Lyrics, &track.Title,
			&track.Artist, &track.Year, &track.Comment,
			&track.Disc, &track.DiscTotal, &track.Track,
			&track.TrackTotal, &track.Duration, &track.SampleRate,
			&track.Bitrate, &track.Sender, &track.CreatorUser,
			&track.Likes, &track.S3Version,
		)
		if err != nil {
			return nil, 0, err
		}
		tracks = append(tracks, track)
	}

	// Get the total count of records (excluding pagination)
	totalRows, countErr := c.GetTotalTrackCount(queryBuilder)
	if countErr != nil {
		return nil, 0, countErr
	}

	return tracks, totalRows, nil
}

// GetTracksByColumns retrieves an track record from the "track" table based on the provided code.
func (c *PgClient) GetTracksByColumns(code, columns string) (*model.Track, error) {
	getTrackByColumns := squirrel.Select("*").From("tracks")
	getTrackByColumns = getTrackByColumns.Where(squirrel.Eq{columns: code})
	getTrackByColumns = getTrackByColumns.PlaceholderFormat(squirrel.Dollar)

	sql, args, err := getTrackByColumns.ToSql()
	if err != nil {
		return nil, err
	}

	// Execute the query and retrieve the results
	rows, errQuery := c.Pool.Query(context.TODO(), sql, args...)
	if errQuery != nil {
		return nil, errQuery
	}
	defer rows.Close()

	var track model.Track
	if !rows.Next() {
		return nil, fmt.Errorf("no records found for code: %s", code)
	}
	err = rows.Scan(
		&track.ID, &track.CreatedAt, &track.UpdatedAt,
		&track.Album, &track.AlbumArtist, &track.Composer,
		&track.Genre, &track.Lyrics, &track.Title,
		&track.Artist, &track.Year, &track.Comment,
		&track.Disc, &track.DiscTotal, &track.Track,
		&track.TrackTotal, &track.Duration, &track.SampleRate,
		&track.Bitrate, &track.Sender, &track.CreatorUser,
		&track.Likes, &track.S3Version,
	)
	if err != nil {
		return nil, err
	}

	return &track, nil
}

// DeleteTracks deletes a single record from the "track" table based on the provided code.
func (c *PgClient) DeleteTracks(code, columns string) error {
	// Create a new instance of squirrel.DeleteBuilder and specify the table name
	deleteBuilder := squirrel.Delete("tracks").PlaceholderFormat(squirrel.Dollar)

	// Add a WHERE condition to specify the record to delete
	deleteBuilder = deleteBuilder.Where(squirrel.Eq{columns: code})
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

// DeleteTracksAll deletes all records from the "track" table.
func (c *PgClient) DeleteTracksAll() error {
	// Create a new instance of squirrel.DeleteBuilder
	deleteBuilder := squirrel.Delete("").From("tracks")

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

// UpdateTracks updates an track record in the "track" table based on the provided code.
func (c *PgClient) UpdateTracks(track *model.Track) error {
	// Create a new instance of squirrel.UpdateBuilder
	updateBuilder := squirrel.Update("tracks")

	// Add SET clauses to specify the columns and their new values
	updateBuilder = updateBuilder.SetMap(map[string]interface{}{
		"created_at":   track.CreatedAt,
		"updated_at":   track.UpdatedAt,
		"album":        track.Album,
		"album_artist": track.AlbumArtist,
		"composer":     track.Composer,
		"genre":        track.Genre,
		"lyrics":       track.Lyrics,
		"title":        track.Title,
		"artist":       track.Artist,
		"year":         track.Year,
		"comment":      track.Comment,
		"disc":         track.Disc,
		"disc_total":   track.DiscTotal,
		"track":        track.Track,
		"track_total":  track.TrackTotal,
		"duration":     track.Duration,
		"sample_rate":  track.SampleRate,
		"bitrate":      track.Bitrate,
		"sender":       track.Sender,
		"likes":        track.Likes,
		"s3Version":    track.S3Version,
	})

	// Add a WHERE condition to identify the record to update based on the provided code
	updateBuilder = updateBuilder.Where(squirrel.Eq{"_id": track.ID})

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

func (c *PgClient) GetAllTracks() ([]model.Track, error) {
	selectBuilder := squirrel.Select("*").
		From("tracks").
		Limit(ChunkSize) // Adjust the limit based on your requirements

	return c.executeSelectQuery(selectBuilder)
}

func (c *PgClient) AddTrackToPlaylist(playlistID, trackID string) error {
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

	// Create a new squirrel.InsertBuilder for the playlist_tracks table
	insertBuilder := squirrel.
		Insert("playlist_tracks").
		Columns("playlist_id", "track_id", "position").
		Values(playlistID, trackID, squirrel.Expr("COALESCE((SELECT MAX(position) FROM playlist_tracks WHERE playlist_id = ?), 0) + 1", playlistID))

	// Generate the SQL query
	query, args, err := insertBuilder.PlaceholderFormat(squirrel.Dollar).ToSql()
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

func (c *PgClient) RemoveTrackFromPlaylist(playlistID, trackID string) error {
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

	// Create a DELETE query using Squirrel to remove the track from the playlist_tracks table
	deleteBuilder := squirrel.
		Delete("playlist_tracks").
		Where(squirrel.Eq{"playlist_id": playlistID, "track_id": trackID})

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

func (c *PgClient) GetAllTracksByPositions(playlistID string) ([]model.Track, error) {
	tx, err := c.Pool.Begin(context.TODO())
	if err != nil {
		return nil, err
	}
	defer func() {
		// Defer the rollback and check for errors
		if rErr := tx.Rollback(context.TODO()); rErr != nil && err == nil {
			err = rErr
		}
	}()

	// Create a query to fetch tracks and their positions
	trackQuery := squirrel.Select("pt.track_id, pt.position, t.*").
		From("playlist_tracks pt").
		Join("tracks t ON pt.track_id = t._id").
		Where(squirrel.Eq{"pt.playlist_id": playlistID}).
		OrderBy("pt.position ASC")

	query, args, err := trackQuery.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	// Execute the query within the transaction
	rows, err := tx.Query(context.TODO(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Create a map to group tracks by playlist ID
	var playlistTracks []model.Track

	for rows.Next() {
		var track model.Track
		var position int64
		var trackPlaylistID string
		if err = rows.Scan(
			//			&trackPlaylistID, &position, &track.ID, &track.CreatedAt,
			//			&track.UpdatedAt, &track.Title, &track.Artist,
			//			&track.Description, &track.Sender, &track.CreatorUser,
			//			&track.Likes, &track.S3Version,

			&trackPlaylistID, &position, &track.ID,
			&track.CreatedAt, &track.UpdatedAt,
			&track.Album, &track.AlbumArtist, &track.Composer,
			&track.Genre, &track.Lyrics, &track.Title,
			&track.Artist, &track.Year, &track.Comment,
			&track.Disc, &track.DiscTotal, &track.Track,
			&track.TrackTotal, &track.Duration, &track.SampleRate,
			&track.Bitrate, &track.Sender, &track.CreatorUser,
			&track.Likes, &track.S3Version,
		); err != nil {
			return nil, err
		}

		playlistTracks = append(playlistTracks, track)
	}

	return playlistTracks, nil
}
