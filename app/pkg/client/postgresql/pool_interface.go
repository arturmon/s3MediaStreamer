package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type PoolInterface interface {
	Acquire(ctx context.Context) (*pgx.Conn, error)
	AcquireFunc(ctx context.Context, f func(*pgx.Conn) error) error
	// ... Add other methods you need
}
