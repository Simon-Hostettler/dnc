package db

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
)

func TestDBPath() string {
	return fmt.Sprintf("%s/dnc_test.db", os.TempDir())
}

func TestDBInstance(dbPath string) (*sqlx.DB, error) {
	if _, err := os.Stat(dbPath); err == nil {
		if err = os.Remove(dbPath); err != nil {
			return nil, err
		}
	}
	if handle, err := Open(dbPath); err != nil {
		return nil, err
	} else {
		return handle, nil
	}
}

func DestroyTestDB(db *sqlx.DB, dbPath string) error {
	if err := db.Close(); err != nil {
		return err
	}
	if err := os.Remove(dbPath); err != nil {
		return err
	}
	return nil
}
