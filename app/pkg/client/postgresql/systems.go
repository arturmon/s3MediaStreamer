package postgresql

import (
	"context"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

// ExecInTransaction executes a SQL query within a transaction.
func (c *PgClient) ExecInTransaction(ctx context.Context, sql string, args ...interface{}) error {
	_, span := otel.Tracer("").Start(ctx, "ExecInTransaction")
	defer span.End()

	// Begin transaction
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Execute the SQL query within the transaction
	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// QueryInTransaction executes a SQL query within a transaction and returns the result.
func (c *PgClient) QueryInTransaction(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	_, span := otel.Tracer("").Start(ctx, "QueryInTransaction")
	defer span.End()

	// Begin transaction
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Execute the SQL query within the transaction
	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		rows.Close()
		return nil, err
	}

	return rows, nil
}
