package db

import (
	"embed"
	"fmt"

	"github.com/jmoiron/sqlx"
	goose "github.com/pressly/goose/v3"
)

// migrations holds SQL migration files embedded into the binary.
//
//go:embed migrations/*.sql
var migrations embed.FS

// Applies all up migrations from the embedded filesystem.
func MigrateUp(db *sqlx.DB) error {
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("sqlite3"); err != nil { // goose uses "sqlite3" dialect name
		return fmt.Errorf("db.MigrateUp: set dialect: %w", err)
	}
	if err := goose.Up(db.DB, "migrations"); err != nil {
		return fmt.Errorf("db.MigrateUp: up: %w", err)
	}
	return nil
}

// Rolls back all migrations
func MigrateDown(db *sqlx.DB) error {
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("db.MigrateDown: set dialect: %w", err)
	}
	if err := goose.Down(db.DB, "migrations"); err != nil {
		return fmt.Errorf("db.MigrateDown: down: %w", err)
	}
	return nil
}
