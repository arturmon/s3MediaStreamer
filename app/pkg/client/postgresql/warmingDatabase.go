package postgresql

import (
	"context"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(connectionString string) error {
	m, err := migrate.New("file://migrations/psql", connectionString)
	if err != nil {
		return err
	}
	defer m.Close()

	// Apply pending migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %v", err)
	}

	return nil
}

func (c *PgClient) TableExists(tableName string) (bool, error) {
	query := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1)"
	var exists bool
	err := c.Pool.QueryRow(context.TODO(), query, tableName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
