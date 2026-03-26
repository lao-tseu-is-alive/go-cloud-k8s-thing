package module

import (
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const defaultSqlDbMigrationsPath = "db/migrations"

// Migrations holds the SQL migration files for the Thing module.
//
//go:embed db/migrations/*.sql
var Migrations embed.FS

// Migrate runs the database migrations for this module.
// It uses golang-migrate with the embedded SQL files.
func Migrate(dbDsn string) error {
	d, err := iofs.New(Migrations, defaultSqlDbMigrationsPath)
	if err != nil {
		return fmt.Errorf("thing module: migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d,
		strings.Replace(dbDsn, "postgres", "pgx5", 1))
	if err != nil {
		return fmt.Errorf("thing module: migration init: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("thing module: migration up: %w", err)
	}
	return nil
}
