package postgres

import (
	"context"
	"fmt"
	"s3MediaStreamer/app/model"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Helper function to parse the path and return the components
func parsePath(pathStr string) (string, string, string, string, uuid.UUID, error) {
	components := strings.Split(pathStr, ".")
	if len(components) != 4 {
		return "", "", "", "", uuid.Nil, fmt.Errorf("invalid path format: %s", pathStr)
	}

	playlistID := components[0]
	trackType := components[1]
	trackID := components[2]

	playlistIDUUID, parseErr := uuid.Parse(playlistID)
	if parseErr != nil {
		return "", "", "", "", uuid.Nil, fmt.Errorf("invalid UUID format for playlistID: %s", playlistID)
	}

	return playlistID, trackType, trackID, pathStr, playlistIDUUID, nil
}

// Helper function to execute SQL update query
func executeUpdateQuery(ctx context.Context, tx pgx.Tx, playlistIDUUID uuid.UUID, oldPath, newPath string) error {
	updateQuery := squirrel.Update("playlist_tracks").
		Set("path", newPath).
		Where(squirrel.Eq{
			"playlist_id": playlistIDUUID,
			"path":        oldPath,
		}).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := updateQuery.ToSql()
	if err != nil {
		return fmt.Errorf("failed to generate update SQL: %w", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute update query: %w", err)
	}

	return nil
}

// UpdatePositionsInDB updates the path with the new position in the database after rebalancing
func (c *Client) UpdatePositionsInDB(ctx context.Context, tree *treemap.Map) error {
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

		playlistID, trackType, trackID, oldPath, playlistIDUUID, parseErr := parsePath(pathStr)
		if parseErr != nil {
			err = parseErr
			return
		}

		newPath := fmt.Sprintf("%s.%s.%s.%d", playlistID, trackType, trackID, node.Position)

		if execErr := executeUpdateQuery(ctx, tx, playlistIDUUID, oldPath, newPath); execErr != nil {
			err = execErr
			return
		}
	})

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// InsertPositionInDB inserts a new position in the database
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

		playlistID, trackType, trackID, _, playlistIDUUID, parseErr := parsePath(pathStr)
		if parseErr != nil {
			err = parseErr
			return
		}

		newPath := fmt.Sprintf("%s.%s.%s.%d", playlistID, trackType, trackID, node.Position)

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

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
