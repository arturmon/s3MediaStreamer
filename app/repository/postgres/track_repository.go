package postgres

import (
	"context"
	"fmt"
	"s3MediaStreamer/app/model"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/jackc/pgx/v5"
)

const ChunkSize = 1000

type TracksRepositoryInterface interface {
	CreateTracks(ctx context.Context, list []model.Track) error
	GetTracks(ctx context.Context, offset, limit int, sortBy, sortOrder, filter, startT, endT string) ([]model.Track, int, error)
	GetTracksByColumns(ctx context.Context, code, columns string) (*model.Track, error)
	CleanTracks(ctx context.Context) error
	DeleteTracksAll(ctx context.Context) error
	UpdateTracks(ctx context.Context, track *model.Track) error
	GetAllTracks(ctx context.Context) ([]model.Track, error)
	AddTrackToPlaylist(ctx context.Context, playlistID, referenceType, referenceID, parentPath string) error
	RemoveTrackFromPlaylist(ctx context.Context, playlistID, trackID string) error
	GetAllTracksByPositions(ctx context.Context, playlistID string) ([]model.Track, error)
	// playlist_tree.go
	UpdatePositionsInDB(ctx context.Context, tree *treemap.Map) error
	InsertPositionInDB(ctx context.Context, tree *treemap.Map) error
	//	ValidateParentPath(ctx context.Context, parentPath, playlistID string) bool
	//	GetPathByReferenceID(ctx context.Context, playlistID, referenceID string) (string, error)
}

// CreateTracks inserts multiple track records into the "track" table.
func (c *Client) CreateTracks(ctx context.Context, list []model.Track) error {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "CreateTracks")
	defer span.End()

	if len(list) == 0 {
		return nil
	}

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

	// Prepare the squirrel insert builder
	ib := squirrel.Insert("tracks").Columns(
		"_id", "created_at", "updated_at", "album", "album_artist",
		"composer", "genre", "lyrics", "title", "artist", "year",
		"comment", "disc", "disc_total", "track", "track_total",
		"duration", "sample_rate", "bitrate",
	)

	// Add INSERT queries to the batch for each track
	for _, track := range list {
		// Build the insert query for each track
		ib = ib.Values(
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
		)
	}
	ib = ib.PlaceholderFormat(squirrel.Dollar)

	// Get the SQL query and arguments from the squirrel builder
	sql, args, err := ib.ToSql()
	if err != nil {
		return err
	}

	// Queue the SQL query and arguments to the batch
	batch := &pgx.Batch{}
	batch.Queue(sql, args...)

	// Execute the batch
	results := c.Pool.SendBatch(ctx, batch)

	// Check for errors in the batch execution
	if err = results.Close(); err != nil {
		return err
	}

	// Commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

// GetTracks retrieves a list of tracks with pagination and filtering.
func (c *Client) GetTracks(
	ctx context.Context,
	offset, limit int,
	sortBy, sortOrder, filter, startT, endT string,
) ([]model.Track, int, error) {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "GetTracks")
	defer span.End()

	queryBuilder := squirrel.Select("*").From("tracks")

	// Apply filtering
	queryBuilder = buildFilterClause(queryBuilder, filter)

	// Apply time-based filtering
	queryBuilder = applyTimeFilters(queryBuilder, startT, endT)

	// Apply sorting
	queryBuilder = buildSortClause(queryBuilder, sortBy, sortOrder)

	// Apply pagination
	queryBuilder = applyPagination(queryBuilder, offset, limit)

	queryBuilder = queryBuilder.PlaceholderFormat(squirrel.Dollar)

	// Generate the SQL query and arguments
	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, 0, err
	}

	// Execute the query
	rows, err := c.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

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
			&track.Bitrate,
		)
		if err != nil {
			return nil, 0, err
		}
		tracks = append(tracks, track)
	}

	// Get total count
	totalRows, countErr := c.GetTotalTrackCount(queryBuilder)
	if countErr != nil {
		return nil, 0, countErr
	}

	return tracks, totalRows, nil
}

// GetTracksByColumns retrieves an track record from the "track" table based on the provided code.
func (c *Client) GetTracksByColumns(ctx context.Context, code, columns string) (*model.Track, error) {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "GetTracksByColumns")
	defer span.End()

	getTrackByColumns := squirrel.Select("*").From("tracks")
	getTrackByColumns = getTrackByColumns.Where(squirrel.Eq{columns: code})
	getTrackByColumns = getTrackByColumns.PlaceholderFormat(squirrel.Dollar)

	sql, args, err := getTrackByColumns.ToSql()
	if err != nil {
		return nil, err
	}

	// Execute the query and retrieve the results
	rows, errQuery := c.Pool.Query(ctx, sql, args...)
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
		&track.Bitrate,
	)
	if err != nil {
		return nil, err
	}

	return &track, nil
}

// CleanTracks deletes a single record from the "track" table based on the provided code.
func (c *Client) CleanTracks(ctx context.Context) error {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "CleanTracks")
	defer span.End()

	// Create a new instance of squirrel.DeleteBuilder and specify the table name
	generateSQLTracks := squirrel.Delete("tracks")

	// Add a WHERE condition to specify the record to delete
	generateSQLTracks = generateSQLTracks.Where("_id NOT IN (SELECT track_id FROM s3Version)")
	generateSQLTracks = generateSQLTracks.PlaceholderFormat(squirrel.Dollar)

	// Generate the SQL query and arguments
	sql, args, err := generateSQLTracks.ToSql()
	if err != nil {
		return err
	}

	return c.ExecInTransaction(ctx, sql, args)
}

// DeleteTracksAll deletes all records from the "track" table.
func (c *Client) DeleteTracksAll(ctx context.Context) error {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "DeleteTracksAll")
	defer span.End()

	// Create a new instance of squirrel.DeleteBuilder
	generateSQLTracks := squirrel.Delete("").From("tracks")

	// Generate the SQL query and arguments
	sql, args, err := generateSQLTracks.ToSql()
	if err != nil {
		return err
	}

	// Execute the DELETE query within a transaction
	return c.ExecInTransaction(ctx, sql, args)
}

// UpdateTracks updates an track record in the "track" table based on the provided code.
func (c *Client) UpdateTracks(ctx context.Context, track *model.Track) error {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "UpdateTracks")
	defer span.End()

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
	})

	// Add a WHERE condition to identify the record to update based on the provided code
	updateBuilder = updateBuilder.Where(squirrel.Eq{"_id": track.ID})

	// Generate the SQL query and arguments
	sql, args, err := updateBuilder.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return err
	}

	// Execute the UPDATE query
	_, err = c.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetAllTracks(ctx context.Context) ([]model.Track, error) {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "GetAllTracks")
	defer span.End()

	selectBuilder := squirrel.Select("*").
		From("tracks").
		Limit(ChunkSize) // Adjust the limit based on your requirements

	return c.ExecuteSelectQuery(ctx, selectBuilder)
}

// AddTrackToPlaylist inserts a track or playlist into a playlist_tracks table supporting nested structures with LTREE.
func (c *Client) AddTrackToPlaylist(ctx context.Context, playlistID, referenceType, referenceID, parentPath string) error {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "AddTrackToPlaylist")
	defer span.End()

	// Start a transaction
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && err == nil {
			err = rErr
		}
	}()

	// Get the maximum position in the current parent path
	var maxPosition int
	positionQuery := squirrel.Select("COALESCE(MAX(CAST(ltree2text(subpath(path, -1, 1)) AS INTEGER)), 0)").
		From("playlist_tracks").
		Where("path <@ ?", parentPath).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := positionQuery.ToSql()
	if err != nil {
		return err
	}

	err = tx.QueryRow(ctx, sql, args...).Scan(&maxPosition)
	if err != nil {
		return err
	}

	// Increment the position for the new item
	newPosition := maxPosition + 1

	// Compute the new path with the updated position
	newPath := fmt.Sprintf("%s.%s.%s.%d", parentPath, referenceType, referenceID, newPosition)

	// Insert into the playlist_tracks table
	insertBuilder := squirrel.Insert("playlist_tracks").
		Columns("playlist_id", "path").
		Values(playlistID, newPath).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

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

// RemoveTrackFromPlaylist removes a track from a playlist_tracks table, handling hierarchical relationships with LTREE.
func (c *Client) RemoveTrackFromPlaylist(ctx context.Context, playlistID, trackID string) error {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "RemoveTrackFromPlaylist")
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

	// Create a DELETE query using LTREE to remove the track and its nested paths
	deleteBuilder := squirrel.
		Delete("playlist_tracks").
		Where(squirrel.Expr(
			"path <@ (SELECT path FROM playlist_tracks WHERE playlist_id = ? AND ltree2text(path) LIKE ?)",
			playlistID,
			"%."+trackID+".%",
		)).
		PlaceholderFormat(squirrel.Dollar)

	// Generate the SQL query and arguments
	query, args, err := deleteBuilder.ToSql()
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

// GetAllTracksByPositions retrieves all tracks within a playlist, including nested ones, ordered by position using LTREE.
func (c *Client) GetAllTracksByPositions(ctx context.Context, playlistID string) ([]model.Track, error) {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "GetAllTracksByPositions")
	defer span.End()

	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		// Defer the rollback and check for errors
		if rErr := tx.Rollback(ctx); rErr != nil && err == nil {
			err = rErr
		}
	}()

	// Construct the path pattern
	pathPattern := playlistID

	// Create a query to fetch tracks ordered by their hierarchical path
	trackQuery := squirrel.Select("t.*").
		From("playlist_tracks pt").
		Join("tracks t ON t._id = pt.path").
		Where(squirrel.Expr("pt.path <@ ?", pathPattern)). // Fetch all tracks in the hierarchy
		OrderBy("pt.path ASC").
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := trackQuery.ToSql()
	if err != nil {
		return nil, err
	}

	// Execute the query
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process the result
	var playlistTracks []model.Track
	for rows.Next() {
		var track model.Track
		err = rows.Scan(
			&track.ID, &track.CreatedAt, &track.UpdatedAt,
			&track.Album, &track.AlbumArtist, &track.Composer,
			&track.Genre, &track.Lyrics, &track.Title,
			&track.Artist, &track.Year, &track.Comment,
			&track.Disc, &track.DiscTotal, &track.Track,
			&track.TrackTotal, &track.Duration, &track.SampleRate,
			&track.Bitrate,
		)
		if err != nil {
			return nil, err
		}
		playlistTracks = append(playlistTracks, track)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return playlistTracks, nil
}

// Helper function to build the filter clause.
func buildFilterClause(queryBuilder squirrel.SelectBuilder, filter string) squirrel.SelectBuilder {
	if filter == "" {
		return queryBuilder
	}

	filterColumns := []string{"album_artist", "composer", "artist"}
	var filterExprs []string

	if strings.HasPrefix(filter, "=") {
		filter = strings.TrimPrefix(filter, "=")
		for _, col := range filterColumns {
			filterExpr := fmt.Sprintf("%s = $%d", col, 1)
			filterExprs = append(filterExprs, filterExpr)
		}
	} else {
		filter = "%" + filter + "%"
		for _, col := range filterColumns {
			filterExpr := fmt.Sprintf("%s ILIKE $%d", col, 1)
			filterExprs = append(filterExprs, filterExpr)
		}
	}

	orCondition := strings.Join(filterExprs, " OR ")
	queryBuilder = queryBuilder.Where(orCondition, filter)
	return queryBuilder
}

// Helper function to apply time filters.
func applyTimeFilters(queryBuilder squirrel.SelectBuilder, startT, endT string) squirrel.SelectBuilder {
	if startT != "" {
		queryBuilder = queryBuilder.Where("updated_at >= $1", startT)
	}
	if endT != "" {
		queryBuilder = queryBuilder.Where("updated_at <= $2", endT)
	}
	return queryBuilder
}

// Helper function to apply sorting.
func buildSortClause(queryBuilder squirrel.SelectBuilder, sortBy, sortOrder string) squirrel.SelectBuilder {
	// Define allowed sort columns and order directions
	allowedSortColumns := map[string]bool{
		"created_at":   true,
		"updated_at":   true,
		"album":        true,
		"album_artist": true,
		"title":        true,
	}
	allowedSortOrders := map[string]bool{
		"ASC":  true,
		"DESC": true,
	}

	// Validate sortBy and sortOrder against allowed values
	if allowedSortColumns[sortBy] && allowedSortOrders[sortOrder] {
		queryBuilder = queryBuilder.OrderBy(fmt.Sprintf("%s %s", sortBy, sortOrder))
	}
	return queryBuilder
}

// Helper function to apply pagination.
func applyPagination(queryBuilder squirrel.SelectBuilder, offset, limit int) squirrel.SelectBuilder {
	if limit < 0 || offset < 0 {
		limit = 0
		offset = 0
	}
	queryBuilder = queryBuilder.Limit(uint64(limit)).Offset(uint64(offset))
	return queryBuilder
}
