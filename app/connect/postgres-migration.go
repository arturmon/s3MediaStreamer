package connect

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" //lint:ignore blank-imports
	_ "github.com/golang-migrate/migrate/v4/source/file"       //lint:ignore blank-imports
)

func RunMigrations(_ context.Context, connectionString string) error {
	m, err := migrate.New("file://migrations/psql", connectionString)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}
	defer func(m *migrate.Migrate) {
		err, _ = m.Close()
		if err != nil {

		}
	}(m)

	// Apply pending migrations
	migrateErr := m.Up()
	if migrateErr != nil {
		if !errors.Is(migrateErr, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations: %w", migrateErr)
		}
	}
	return nil
}
