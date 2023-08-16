package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" //lint:ignore blank-imports
	_ "github.com/golang-migrate/migrate/v4/source/file"       //lint:ignore blank-imports
)

func RunMigrations(connectionString string) error {
	m, err := migrate.New("file://migrations/psql", connectionString)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}
	defer m.Close()

	// Apply pending migrations
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}
	}

	return nil
}

func (c *PgClient) TableExists(tableName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = current_schema() AND table_name = $1
		)`
	var exists bool
	err := c.Pool.QueryRow(context.TODO(), query, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}
	return exists, nil
}
