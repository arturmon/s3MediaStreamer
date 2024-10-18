package postgres

import (
	"context"
	"fmt"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"

	"github.com/Masterminds/squirrel"
)

type UtilsRepositoryInterface interface {
	ExecuteSelectQuery(ctx context.Context, selectBuilder squirrel.SelectBuilder) ([]model.Track, error)
	Connect(_ *logs.Logger) error
	Ping(ctx context.Context) error
	Close(_ context.Context) error
	ExecInTransaction(ctx context.Context, sql string, args ...interface{}) error
	QueryInTransaction(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	GetConnectionString() string
	Begin(ctx context.Context) (pgx.Tx, error)
}

func (c *Client) ExecuteSelectQuery(ctx context.Context, selectBuilder squirrel.SelectBuilder) ([]model.Track, error) {
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
		rows, err := c.Pool.Query(ctx, sql, args...)
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

func (c *Client) Connect(_ *logs.Logger) error {
	if c.Pool != nil {
		conn, connErr := c.Pool.Acquire(context.Background())
		if connErr != nil {
			return connErr
		}
		defer conn.Release()
		if pingErr := conn.Conn().Ping(context.Background()); pingErr != nil {
			return pingErr
		}
	} else {
		return fmt.Errorf("pgx pool is not initialized")
	}
	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	if c.Pool != nil {
		conn, connErr := c.Pool.Acquire(ctx)
		if connErr != nil {
			return connErr
		}
		defer conn.Release()
		pingErr := conn.Conn().Ping(ctx)
		if pingErr != nil {
			return pingErr
		}
	} else {
		return fmt.Errorf("pgx pool is not initialized")
	}
	return nil
}

func (c *Client) Close(_ context.Context) error {
	if c.Pool != nil {
		c.Pool.Close()
		c.Pool = nil
	}
	return nil
}

// ExecInTransaction executes a SQL query within a transaction.
func (c *Client) ExecInTransaction(ctx context.Context, sql string, args []interface{}) error {
	_, span := otel.Tracer("").Start(ctx, "ExecInTransaction")
	defer span.End()

	// Begin transaction
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// Check and handle the error from Rollback
		if rErr := tx.Rollback(ctx); rErr != nil && err == nil {
			err = rErr // Only update err if it's nil, prioritizing the original error
		}
	}()

	// Execute the SQL query within the transaction
	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit(ctx)
}

// QueryInTransaction executes a SQL query within a transaction and returns the result.
func (c *Client) QueryInTransaction(ctx context.Context, sql string, args []interface{}) (pgx.Rows, error) {
	_, span := otel.Tracer("").Start(ctx, "QueryInTransaction")
	defer span.End()

	// Begin transaction
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		// Check and handle the error from Rollback
		if rErr := tx.Rollback(ctx); rErr != nil && err == nil {
			err = rErr // Only update err if it's nil, prioritizing the original error
		}
	}()

	// Execute the SQL query within the transaction
	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return rows, nil
}

func (c *Client) GetConnectionString() string {
	return c.ConnectionString
}

func (c *Client) Begin(ctx context.Context) (pgx.Tx, error) {
	return c.Pool.Begin(ctx)
}
