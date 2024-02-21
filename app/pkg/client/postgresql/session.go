package postgresql

import (
	"context"
	"github.com/Masterminds/squirrel"
)

func (c *PgClient) CleanSessions() error {
	// Build the SQL query using squirrel
	condition := squirrel.Delete("http_sessions").
		Where("expires_on < NOW()").
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := condition.ToSql()
	if err != nil {
		return err
	}

	// Execute the SQL query
	_, err = c.Pool.Exec(context.TODO(), query, args...)
	return err
}
