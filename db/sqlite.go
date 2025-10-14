package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// Opens (and creates if missing) SQLite db at provided path
func Open(dbPath string) (*sqlx.DB, error) {
	if dbPath == "" {
		return nil, errors.New("db.Open: empty database path")
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("db.Open: ensure dir: %w", err)
	}

	// Enable foreign keys via pragmas. modernc supports DSN query params.
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", filepath.ToSlash(dbPath))
	db, err := sqlx.Open("sqlite", dsn) // modernc driver named sqlite
	if err != nil {
		return nil, fmt.Errorf("db.Open: open: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err := ping(db.DB); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func ping(sdb *sql.DB) error {
	if err := sdb.Ping(); err != nil {
		return fmt.Errorf("db.Open: ping: %w", err)
	}
	// Ensure foreign keys are really on for current connection.
	if _, err := sdb.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("db.Open: enable foreign_keys: %w", err)
	}
	return nil
}
