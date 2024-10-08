package postgres

import (
	"context"
	"errors"
	"fmt"
	"s3MediaStreamer/app/model"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PlaylistRepositoryInterface interface {
	CheckPlaylistExists(ctx context.Context, playlistID string) (bool, error)
	CreatePlayListName(ctx context.Context, playlist model.PLayList) error
	FetchPlaylistInfo(ctx context.Context, playlistID string, playlist *model.PLayList) error
	GetTracksByPlaylist(ctx context.Context, playlistID string) ([]model.Track, error)
	ClearPlaylistContents(ctx context.Context, playlistID string) error
	DeletePlaylist(ctx context.Context, playlistID string) error
	UpdatePlaylistDetails(ctx context.Context, playlistID, title, description string) error
	GetPlaylistOwner(ctx context.Context, playlistID string) (uuid.UUID, error)
	GetPlaylists(ctx context.Context, userID string) ([]model.PLayList, error)
	GetPlaylistAllTracks(ctx context.Context, playlistID string) ([]model.TrackRequest, error)
	GetPlaylistItems(ctx context.Context, playlistID string) ([]model.PlaylistStruct, error)
	GetPlaylistPath(ctx context.Context, playlistID string) (string, error)
}

// CheckPlaylistExists checks whether a playlist exists in the database based on its ID.
// It counts the entries in the "playlists" table with the provided playlist ID.
//
// Parameters:
//   - ctx: context.Context
//     A context that carries deadlines, cancellation signals, and other request-scoped values.
//     It helps to control the lifecycle of the function execution.
//   - playlistID: string
//     The unique identifier of the playlist to be checked.
//
// Return Values:
//   - bool: Returns true if the playlist exists in the database, false otherwise.
//   - error: Returns an error if any issues occur during SQL query execution or data retrieval.
func (c *Client) CheckPlaylistExists(ctx context.Context, playlistID string) (bool, error) {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "CheckPlaylistExists")
	defer span.End()

	queryBuilder := squirrel.Select("COUNT(*)").From("playlists").Where(squirrel.Eq{"_id": playlistID}).PlaceholderFormat(squirrel.Dollar)
	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return false, err
	}

	var count int
	err = c.Pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil || count == 0 {
		return false, errors.New("playlist does not exist")
	}
	return count > 0, nil
}

// CreatePlayListName creates a new playlist entry in the "playlists" table.
// It ensures atomicity by executing the insertion within a transaction.
//
// Parameters:
//   - ctx: context.Context
//     A context that provides a way to handle timeouts, cancellation signals, and tracing (via OpenTelemetry).
//   - playlist: model.PlayList
//     A playlist object containing the playlist's information, such as ID, title, description, and creator's user ID.
//
// Return Value:
//   - error: Returns an error if the transaction or SQL execution fails; otherwise, returns nil.
func (c *Client) CreatePlayListName(ctx context.Context, playlist model.PLayList) error {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "CreatePlayListName")
	defer span.End()

	return c.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Create a new squirrel.InsertBuilder
		insertBuilder := squirrel.
			Insert("playlists").
			Columns("_id", "created_at", "title", "description", "_creator_user").
			Values(
				playlist.ID,
				playlist.CreatedAt,
				playlist.Title,
				playlist.Description,
				playlist.CreatorUser,
			).PlaceholderFormat(squirrel.Dollar)

		return ExecuteSQL(ctx, tx, insertBuilder)
	})
}

// FetchPlaylistInfo fetches basic information about a playlist from the "playlists" table.
//
// Parameters:
//   - ctx: context.Context
//     A context that provides cancellation signals and tracing functionality.
//   - playlistID: string
//     The unique identifier for the playlist whose information is to be fetched.
//   - playlist: *model.PLayList
//     A pointer to a playlist struct that will be populated with the retrieved playlist details.
//
// Return Value:
//   - error: Returns an error if the SQL query or data retrieval fails, otherwise returns nil.
func (c *Client) FetchPlaylistInfo(ctx context.Context, playlistID string, playlist *model.PLayList) error {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "FetchPlaylistInfo")
	defer span.End()

	selectQuery := squirrel.Select("*").From("playlists").Where(squirrel.Eq{"_id": playlistID}).PlaceholderFormat(squirrel.Dollar)
	sql, args, err := selectQuery.ToSql()
	if err != nil {
		return err
	}

	return c.Pool.QueryRow(ctx, sql, args...).Scan(
		&playlist.ID,
		&playlist.CreatedAt,
		&playlist.Title,
		&playlist.Description,
		&playlist.CreatorUser,
	)
}

// GetTracksByPlaylist fetches all tracks associated with a specific playlist.
// It joins the "tracks" table with the "playlist_tracks" table to return tracks linked to the playlist.
//
// Parameters:
//   - ctx: context.Context
//     The context manages timeouts and tracing.
//   - playlistID: string
//     The ID of the playlist for which tracks are to be fetched.
//
// Return Values:
//   - []model.Track: A slice of tracks associated with the playlist. Each track contains metadata such as title, artist, etc.
//   - error: Returns an error if any issues occur during SQL query execution or row scanning.
func (c *Client) GetTracksByPlaylist(ctx context.Context, playlistID string) ([]model.Track, error) {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "GetTracksByPlaylist")
	defer span.End()

	selectQuery := squirrel.Select("t.*").
		From("tracks t").
		Join("playlist_tracks pt ON t._id = pt.path").
		Where(squirrel.Expr("pt.path ~ ?", playlistID+".track.*")).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, err
	}

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

// ClearPlaylistContents removes all tracks from the specified playlist but keeps the playlist itself intact.
//
// Parameters:
//   - ctx: context.Context
//     The context that controls the request lifecycle and provides cancellation support.
//   - playlistID: string
//     The unique identifier of the playlist whose tracks are to be cleared.
//
// Return Value:
//   - error: Returns an error if the SQL query execution or transaction fails.
func (c *Client) ClearPlaylistContents(ctx context.Context, playlistID string) error {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "ClearPlaylistContents")
	defer span.End()

	return c.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		clearQuery := squirrel.Delete("playlist_tracks").
			Where(squirrel.Expr("path <@ ?", playlistID)).
			PlaceholderFormat(squirrel.Dollar)

		if err := ExecuteSQL(ctx, tx, clearQuery); err != nil {
			return err
		}

		return nil
	})
}

// DeletePlaylist removes a playlist and all its associated tracks from the database.
//
// Parameters:
//   - ctx: context.Context
//     A context that provides control over timeouts and tracing during the operation.
//   - playlistID: string
//     The unique identifier for the playlist to be deleted.
//
// Return Value:
//   - error: Returns an error if the SQL query execution or transaction fails.
func (c *Client) DeletePlaylist(ctx context.Context, playlistID string) error {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "DeletePlaylist")
	defer span.End()

	return c.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// First, delete all associated tracks from playlist_tracks
		deleteTracksQuery := squirrel.Delete("playlist_tracks").
			Where(squirrel.Expr("path <@ ?", playlistID)).
			PlaceholderFormat(squirrel.Dollar)

		if err := ExecuteSQL(ctx, tx, deleteTracksQuery); err != nil {
			return err
		}

		// Then, delete the playlist itself from the playlists table
		deletePlaylistQuery := squirrel.Delete("playlists").
			Where(squirrel.Eq{"_id": playlistID}).
			PlaceholderFormat(squirrel.Dollar)

		if err := ExecuteSQL(ctx, tx, deletePlaylistQuery); err != nil {
			return err
		}

		return nil
	})
}

// UpdatePlaylistDetails updates the title and description of a playlist in the "playlists" table.
// The function ensures that the operation is executed within a transaction for atomicity.
//
// Parameters:
//   - ctx: context.Context
//     A context for managing the request lifecycle, including timeouts and tracing.
//   - playlistID: string
//     The unique identifier of the playlist to be updated.
//   - title: string
//     The new title for the playlist. Pass an empty string to retain the current title.
//   - description: string
//     The new description for the playlist. Pass an empty string to retain the current description.
//
// Return Value:
//   - error: Returns an error if the SQL update query or transaction fails; otherwise, returns nil.
func (c *Client) UpdatePlaylistDetails(ctx context.Context, playlistID, title, description string) error {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "UpdatePlaylistDetails")
	defer span.End()

	return c.ExecuteInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Build the update query
		updateQuery := squirrel.Update("playlists").
			Set("title", title).
			Set("description", description).
			Where(squirrel.Eq{"_id": playlistID}).
			PlaceholderFormat(squirrel.Dollar)

		return ExecuteSQL(ctx, tx, updateQuery)
	})
}

// GetRootPlaylistID retrieves the ID of the root playlist from the database.
// The root playlist is defined as a playlist without a parent playlist (i.e., parent_id is NULL).
//
// Parameters:
//   - ctx: context.Context
//     A context that carries deadlines, cancellation signals, and other request-scoped values.
//
// Return Values:
//   - string: The ID of the root playlist if it exists.
//   - error: Returns an error if there is an issue executing the SQL query or retrieving the data.
func (c *Client) GetRootPlaylistID(ctx context.Context) (string, error) {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "GetRootPlaylistID")
	defer span.End()

	queryBuilder := squirrel.Select("_id").
		From("playlists").
		Where(squirrel.Eq{"parent_id": nil}). // Playlist without a parent
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return "", err
	}

	var rootPlaylistID string
	err = c.Pool.QueryRow(ctx, sql, args...).Scan(&rootPlaylistID)
	if err != nil {
		return "", err
	}

	return rootPlaylistID, nil
}

// GetPlaylistOwner retrieves the username of the owner of a specified playlist.
//
// Parameters:
//   - ctx: context.Context
//     The context for managing request deadlines, cancellation signals, and tracing.
//   - playlistID: string
//     The unique identifier of the playlist for which the owner's information is to be retrieved.
//
// Return Values:
//   - string: The username of the playlist's creator if found.
//   - error: Returns an error if there is an issue fetching the playlist info or if the playlist does not exist.
func (c *Client) GetPlaylistOwner(ctx context.Context, playlistID string) (uuid.UUID, error) {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "GetPlaylistOwner")
	defer span.End()

	var playlist model.PLayList
	// Fetch basic playlist info
	if err := c.FetchPlaylistInfo(ctx, playlistID, &playlist); err != nil {
		return uuid.Nil, err
	}

	creatorUser := playlist.CreatorUser.String()

	uuidValue, err := uuid.Parse(creatorUser)
	if err != nil {
		return uuid.Nil, err
	}
	return uuidValue, nil
}

// GetRootPlaylistOwner retrieves the username of the owner of the root playlist.
// It first retrieves the ID of the root playlist and then fetches the owner's information.
//
// Parameters:
//   - ctx: context.Context
//     The context for managing request deadlines, cancellation signals, and tracing.
//
// Return Values:
//   - string: The username of the owner of the root playlist.
//   - error: Returns an error if there is an issue retrieving the root playlist ID or owner information.
func (c *Client) GetRootPlaylistOwner(ctx context.Context) (uuid.UUID, error) {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "GetRootPlaylistOwner")
	defer span.End()

	rootPlaylistID, err := c.GetRootPlaylistID(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	ownerID, err := c.GetPlaylistOwner(ctx, rootPlaylistID)
	if err != nil {
		return uuid.Nil, err
	}

	return ownerID, nil
}

// GetPlaylists fetches playlists based on the userID. If userID is provided, it fetches the playlists of that user.
// If userID is empty, it fetches all playlists in the database.
//
// Parameters:
//   - ctx: context.Context
//     A context that handles the request lifecycle and deadlines.
//   - userID: uuid.UUID or string
//     The unique identifier of the user whose playlists we want to fetch. Pass an empty string to fetch all playlists.
//
// Return Values:
//   - []model.PlayList: A slice of playlists. Each playlist contains metadata such as ID, title, description, etc.
//   - error: Returns an error if the SQL query execution or row scanning fails.
func (c *Client) GetPlaylists(ctx context.Context, userID string) ([]model.PLayList, error) {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "GetPlaylists")
	defer span.End()

	// Start building the query to fetch playlists
	queryBuilder := squirrel.Select("*").From("playlists").PlaceholderFormat(squirrel.Dollar)

	// If userID is provided, filter by userID
	if userID != "" {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"_creator_user": userID})
	}

	// Convert query to SQL
	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	// Execute the query
	rows, err := c.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize an empty slice to hold the playlists
	var playlists []model.PLayList

	// Loop through the result rows
	for rows.Next() {
		var playlist model.PLayList
		err = rows.Scan(
			&playlist.ID,
			&playlist.CreatedAt,
			&playlist.Title,
			&playlist.Description,
			&playlist.CreatorUser,
		)
		if err != nil {
			return nil, err
		}
		// Append each playlist to the slice
		playlists = append(playlists, playlist)
	}

	// Return the slice of playlists
	return playlists, nil
}

// GetPlaylistAllTracks fetches a playlist, its tracks, and its nested playlists and tracks in a single query using Squirrel.
//
// Parameters:
//   - ctx: context.Context
//     The context that carries deadlines, cancellation signals, and other request-scoped values.
//   - playlistID: string
//     The unique identifier of the root playlist.
//
// Return Values:
//   - []model.TrackRequest: A slice that includes the tracks in the playlist and nested playlists.
//   - error: Returns an error if there is an issue executing the SQL queries or retrieving the data.
func (c *Client) GetPlaylistAllTracks(ctx context.Context, playlistID string) ([]model.TrackRequest, error) {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "GetPlaylistAllTracks")
	defer span.End()

	selectQuery := squirrel.Expr(`
		SELECT t.*, 
		       subpath(pt.path, 0, 1)::text AS playlist_id,  -- Extract the first element as playlist_id
		       subpath(pt.path, nlevel(pt.path) - 1, 1)::text AS position  -- Extract the last element as position
		FROM tracks t
		JOIN playlist_tracks pt
		     ON subpath(pt.path, nlevel(pt.path) - 2, 1)::text::uuid = t._id  -- Extract the track UUID
		JOIN (SELECT path FROM playlist_tracks WHERE playlist_id = $1) p_filter
		     ON pt.path <@ p_filter.path
	`, playlistID)

	sql, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := c.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//var tracks []model.TrackRequest
	/*
		//TODO:
		var results []map[string]interface{}

		for rows.Next() {
			// Получаем описание полей
			fieldDescriptions := rows.FieldDescriptions()
			values := make([]interface{}, len(fieldDescriptions))
			for i := range values {
				values[i] = new(interface{}) // создаем указатели для хранения значений
			}

			// Считываем значения в переменную values
			if err := rows.Scan(values...); err != nil {
				return nil, err
			}

			row := make(map[string]interface{})
			for i, field := range fieldDescriptions {
				val := *(values[i].(*interface{}))
				row[string(field.Name)] = val // Преобразуем []byte в строку
			}
			results = append(results, row)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}

		for _, result := range results {
			fmt.Println(result)
		}

	*/
	var trackRequests []model.TrackRequest

	for rows.Next() {
		var track model.Track
		var readPlaylistID, position string // Переменная для хранения позиции

		// Считывание данных в переменные
		if err = rows.Scan(
			&track.ID, &track.CreatedAt, &track.UpdatedAt,
			&track.Album, &track.AlbumArtist, &track.Composer,
			&track.Genre, &track.Lyrics, &track.Title,
			&track.Artist, &track.Year, &track.Comment,
			&track.Disc, &track.DiscTotal, &track.Track,
			&track.TrackTotal, &track.Duration, &track.SampleRate,
			&track.Bitrate,
			&readPlaylistID, // Здесь считываем позицию
			&position,       // И здесь считываем PlaylistID
		); err != nil {
			return nil, err
		}

		// Заполняем TrackRequest
		trackRequest := model.TrackRequest{
			Position:   position,
			PlaylistID: readPlaylistID,
			Track:      track,
		}
		trackRequests = append(trackRequests, trackRequest)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Вывод результата
	for _, trackRequest := range trackRequests {
		fmt.Printf("%+v\n", trackRequest)
	}
	tracks := trackRequests

	return tracks, nil
}

// getTracksWithPosition recursively fetches the tracks from a playlist and handles the position incrementing
// for tracks inside both the root and nested playlists.
//
// Parameters:
//   - ctx: context.Context
//     The context that carries deadlines, cancellation signals, and other request-scoped values.
//   - playlistID: string
//     The unique identifier of the current playlist (either root or nested).
//   - lastPosition: int
//     The last known position from the parent playlist. Used as the base position for nested tracks.
//
// Return Values:
//   - []model.TrackRequest: A slice that includes tracks from both the playlist and any nested playlists.
//   - int: The updated position after processing all the tracks.
//   - error: Returns an error if there is an issue executing the SQL queries or retrieving the data.
func (c *Client) getTracksWithPosition(ctx context.Context, playlistID string, position int) ([]model.TrackRequest, int, error) {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "getTracksWithPosition")
	defer span.End()

	var trackRequests []model.TrackRequest

	// Get all playlist items (both tracks and nested playlists)
	_, err := c.GetPlaylistItems(ctx, playlistID)
	if err != nil {
		return nil, position, err
	}
	/*
		for _, item := range items {
			if item.ReferenceType == "track" {
				// Fetch the track data from the `tracks` table
				trackRequest, err := c.getTrackByID(ctx, item.ReferenceID)
				if err != nil {
					return nil, position, err
				}
				trackRequest.Position = position // Set current position (parent playlist)
				trackRequest.PlaylistID, _ = uuid.Parse(playlistID)
				trackRequests = append(trackRequests, trackRequest)
				position++ // Increment position for the next track
			} else if item.ReferenceType == "playlist" {
				// Recursively fetch tracks from the nested playlist
				// Pass the current position to start with nested playlist tracks
				nestedTracks, updatedPosition, err := c.getTracksWithPosition(ctx, item.ReferenceID.String(), position)
				if err != nil {
					return nil, position, err
				}

				trackRequests = append(trackRequests, nestedTracks...)
				// Update the position after processing nested playlist
				position = updatedPosition
			}
		}

	*/

	return trackRequests, position, nil
}

// getTrackByID fetches a single track by its unique identifier and maps it to TrackRequest.
//
// Parameters:
//   - ctx: context.Context
//     The context that carries deadlines, cancellation signals, and other request-scoped values.
//   - trackID: uuid.UUID
//     The unique identifier of the track to be retrieved.
//
// Return Values:
//   - model.TrackRequest: A TrackRequest struct that holds the track's metadata (e.g., title, artist, album, etc.).
//   - error: Returns an error if there is an issue executing the SQL queries or retrieving the data.
func (c *Client) getTrackByID(ctx context.Context, trackID uuid.UUID) (model.TrackRequest, error) {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "getTrackByID")
	defer span.End()

	var trackRequest model.TrackRequest

	// Query to fetch track data
	query := squirrel.Select("_id", "title", "artist", "album", "duration", "album_artist", "composer", "genre", "lyrics", "year", "comment", "disc", "disc_total", "track", "track_total", "sample_rate", "bitrate").
		From("tracks").
		Where("_id = ?", trackID).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return trackRequest, err
	}

	// Execute the query and scan results
	err = c.Pool.QueryRow(ctx, sql, args...).Scan(
		&trackRequest.ID,
		&trackRequest.Title,
		&trackRequest.Artist,
		&trackRequest.Album,
		&trackRequest.Duration,
		&trackRequest.AlbumArtist,
		&trackRequest.Composer,
		&trackRequest.Genre,
		&trackRequest.Lyrics,
		&trackRequest.Year,
		&trackRequest.Comment,
		&trackRequest.Disc,
		&trackRequest.DiscTotal,
		&trackRequest.Track,
		&trackRequest.TrackTotal,
		&trackRequest.SampleRate,
		&trackRequest.Bitrate,
	)
	if err != nil {
		return trackRequest, err
	}

	return trackRequest, nil
}

// GetPlaylistItems retrieves the items (tracks and nested playlists) from a given playlist.
//
// Parameters:
//   - ctx: context.Context
//     The context that carries deadlines, cancellation signals, and other request-scoped values.
//   - playlistID: string
//     The unique identifier of the playlist.
//
// Return Values:
//   - []model.PlaylistStruct: A slice of PlaylistStruct that contains both tracks and nested playlists from the playlist.
//   - error: Returns an error if there is an issue executing the SQL queries or retrieving the data.
func (c *Client) GetPlaylistItems(ctx context.Context, playlistID string) ([]model.PlaylistStruct, error) {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "getPlaylistItems")
	defer span.End()

	var items []model.PlaylistStruct

	query := squirrel.Select("path").
		From("playlist_tracks").
		Where(squirrel.Expr("path <@ ?", playlistID)).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := c.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.PlaylistStruct
		if err = rows.Scan(&item.Path); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// GetPlaylistPath retrieves the LTREE path for a given playlist ID from the playlist_tracks table.
func (c *Client) GetPlaylistPath(ctx context.Context, playlistID string) (string, error) {
	// Get the tracer for the current context
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "GetPlaylistPath")
	defer span.End()

	var path string

	// Build the SQL query to fetch the path from the playlist_tracks table
	query := squirrel.Select("path").
		From("playlist_tracks").
		Where(squirrel.Eq{"playlist_id": playlistID}).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return "", err
	}

	// Execute the query and retrieve the path
	err = c.Pool.QueryRow(ctx, sql, args...).Scan(&path)
	if err != nil {
		return "", err
	}

	return path, nil
}
