package postgresql

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type PoolInterface interface {
	Acquire(ctx context.Context) (*pgx.Conn, error)
	AcquireFunc(ctx context.Context, f func(*pgx.Conn) error) error
	// ... Add other methods you need
}
