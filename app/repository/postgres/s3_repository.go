package postgres

import (
	"errors"

	"context"

	"github.com/Masterminds/squirrel"
)

type S3RepositoryInterface interface {
	GetS3VersionByTrackID(ctx context.Context, trackID string) (string, error)
	AddS3Version(ctx context.Context, trackID, version string) error
	DeleteS3Version(ctx context.Context, version string) error
}

func (c *Client) GetS3VersionByTrackID(ctx context.Context, trackID string) (string, error) {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "GetS3VersionByTrackID")
	defer span.End()

	var version string

	// Create a SQL query to fetch
	selectQuery := squirrel.Select("version").
		From("s3Version").
		Where(squirrel.Eq{"track_id": trackID}).
		PlaceholderFormat(squirrel.Dollar)

	// Convert the SQL query to SQL and arguments
	sql, args, err := selectQuery.ToSql()
	if err != nil {
		return "", err
	}

	rows, err := c.Pool.Query(ctx, sql, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if !rows.Next() {
		return "", errors.New("no connection was found in s3 with this identifier")
	}
	if err = rows.Scan(&version); err != nil {
		return "", err
	}

	return version, nil
}

func (c *Client) AddS3Version(ctx context.Context, trackID, version string) error {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "AddS3Version")
	defer span.End()

	// Create an insert query using Squirrel
	insertQuery := squirrel.Insert("s3Version").
		Columns("track_id", "version").
		Values(trackID, version).
		PlaceholderFormat(squirrel.Dollar)

	// Convert the insert query to SQL and arguments
	sql, args, err := insertQuery.ToSql()
	if err != nil {
		return err
	}

	_, err = c.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteS3Version(ctx context.Context, version string) error {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "DeleteS3Version")
	defer span.End()

	// Create a delete query using Squirrel
	deleteQuery := squirrel.Delete("s3Version").
		Where(squirrel.Eq{"version": version}).
		PlaceholderFormat(squirrel.Dollar)

	// Convert the delete query to SQL and arguments
	sql, args, err := deleteQuery.ToSql()
	if err != nil {
		return err
	}

	_, err = c.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}
