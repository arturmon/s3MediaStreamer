package postgres

import (
	"context"
	"fmt"
	"s3MediaStreamer/app/model"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/google/uuid"
)

// UpdatePositionsInDB updates the path with the new position in the database after rebalancing
func (c *Client) UpdatePositionsInDB(ctx context.Context, tree *treemap.Map) error {

	// Start a transaction
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && err == nil {
			err = rErr
		}
	}()

	if tree.Size() == 0 {
		return fmt.Errorf("tree is empty, nothing to update")
	}

	// Iterate over the tree and update each node's path in the database
	tree.Each(func(key interface{}, value interface{}) {
		node, ok := value.(*model.Node)
		if !ok {
			err = fmt.Errorf("unexpected value type: %s", value)
			return
		}

		// Convert the key (path) to string and split it into components
		pathStr, ok := key.(string)
		if !ok {
			err = fmt.Errorf("unexpected key type: %T", key)
			return
		}
		components := strings.Split(pathStr, ".")

		// We expect the format <playlistID>.<trackType>.<trackID>.<position>
		if len(components) != 4 {
			err = fmt.Errorf("invalid path format: %s", pathStr)
			if err != nil {
				return
			}

		}

		playlistID := components[0] // Playlist ID
		trackType := components[1]  // 'track' or 'playlist'
		trackID := components[2]    // Track or Playlist ID

		// Generate new path with the updated position
		newPath := fmt.Sprintf("%s.%s.%s.%d", playlistID, trackType, trackID, node.Position)

		playlistIDUUID, parseErr := uuid.Parse(components[0]) // Assuming components[0] is the playlistID
		if parseErr != nil {
			err = fmt.Errorf("invalid UUID format for playlistID: %s", components[0])
			return
		}

		// Build the SQL query using squirrel to update the path
		updateQuery := squirrel.Update("playlist_tracks").
			Set("path", newPath). // Update the full path
			Where(squirrel.Eq{
				"playlist_id": playlistIDUUID,
				"path":        pathStr,
			}).
			PlaceholderFormat(squirrel.Dollar)
		// Universal update
		/*
			upsertQuery := squirrel.Insert("playlist_tracks").
				Columns("playlist_id", "path", "position").
				Values(playlistIDUUID, newPath, node.Position).
				Suffix("ON CONFLICT (playlist_id, path) DO UPDATE SET position = EXCLUDED.position").
				PlaceholderFormat(squirrel.Dollar)
		*/

		// Generate the SQL query and arguments
		query, args, errGenerate := updateQuery.ToSql()
		if errGenerate != nil {
			return
		}

		// Execute the update query
		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			return
		}

	})

	// Commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) InsertPositionInDB(ctx context.Context, tree *treemap.Map) error {
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && err == nil {
			err = rErr
		}
	}()

	if tree.Size() == 0 {
		return fmt.Errorf("tree is empty, nothing to insert")
	}

	tree.Each(func(key interface{}, value interface{}) {
		node, ok := value.(*model.Node)
		if !ok {
			err = fmt.Errorf("unexpected value type: %T", value)
			return
		}

		pathStr, ok := key.(string)
		if !ok {
			err = fmt.Errorf("unexpected key type: %T", key)
			return
		}
		components := strings.Split(pathStr, ".")

		if len(components) != 4 {
			err = fmt.Errorf("invalid path format: %s", pathStr)
			return
		}

		playlistID := components[0]
		trackType := components[1]
		trackID := components[2]

		newPath := fmt.Sprintf("%s.%s.%s.%d", playlistID, trackType, trackID, node.Position)

		playlistIDUUID, parseErr := uuid.Parse(playlistID)
		if parseErr != nil {
			err = fmt.Errorf("invalid UUID format for playlistID: %s", playlistID)
			return
		}

		insertQuery := squirrel.Insert("playlist_tracks").
			Columns("playlist_id", "path").
			Values(playlistIDUUID, newPath).
			PlaceholderFormat(squirrel.Dollar)

		query, args, errGenerate := insertQuery.ToSql()
		if errGenerate != nil {
			err = errGenerate
			return
		}

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			return
		}

	})

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
