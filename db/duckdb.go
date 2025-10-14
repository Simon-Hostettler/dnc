package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/marcboeker/go-duckdb"
)

// Opens (creates if missing) duckdb at given path
func Open(dbPath string) (*sqlx.DB, error) {
	if dbPath == "" {
		return nil, errors.New("db.Open: empty database path")
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("db.Open: ensure dir: %w", err)
	}

	db, err := sqlx.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("db.Open: open duckdb: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxIdleTime(10 * time.Minute)

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
	return nil
}
