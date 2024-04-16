package postgresql

import (
	"errors"

	"go.opentelemetry.io/otel"

	"context"

	"github.com/Masterminds/squirrel"
)

func (c *PgClient) GetS3VersionByTrackID(ctx context.Context, trackID string) (string, error) {
	_, span := otel.Tracer("").Start(ctx, "GetS3VersionByTrackID")
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

	if err != nil {
		return "", err
	}

	if !rows.Next() {
		return "", errors.New("no connection was found in s3 with this identifier")
	}
	if err = rows.Scan(&version); err != nil {
		return "", err
	}

	return version, nil
}

func (c *PgClient) AddS3Version(ctx context.Context, trackID, version string) error {
	_, span := otel.Tracer("").Start(ctx, "AddS3Version")
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

func (c *PgClient) DeleteS3Version(ctx context.Context, version string) error {
	_, span := otel.Tracer("").Start(ctx, "DeleteS3Version")
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
