package db

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

var testDBPath = fmt.Sprintf("%s/dnc_test.db", os.TempDir())

func testDBInstance() (*sqlx.DB, error) {
	if _, err := os.Stat(testDBPath); err == nil {
		if err = os.Remove(testDBPath); err != nil {
			return nil, err
		}
	}
	if handle, err := Open(testDBPath); err != nil {
		return nil, err
	} else {
		return handle, nil
	}
}

func destroyTestDB(db *sqlx.DB) error {
	if err := db.Close(); err != nil {
		return err
	}
	if err := os.Remove(testDBPath); err != nil {
		return err
	}
	return nil
}

func TestEmptyMigrations(t *testing.T) {
	handle, err := testDBInstance()
	if err != nil {
		t.Fatalf("Could not create test DB: %s", err.Error())
	}
	if err := MigrateUp(handle); err != nil {
		t.Errorf("Migration to current version failed: %s", err.Error())
	}
	if err := MigrateDown(handle); err != nil {
		t.Errorf("Migrating down to initial DB failed: %s", err.Error())
	}
	if err := destroyTestDB(handle); err != nil {
		t.Fatalf("Could not destroy test DB: %s", err.Error())
	}
}

func TestMigrationSectionsDetected(t *testing.T) {
	testSQL := "-- +duckUp\n" +
		"INSERT INTO students(first, last) VALUES ('Bobby', 'Tables');\n" +
		"-- +duckDown\n" +
		"DROP TABLE students;\n"

	up, err := extractUp(testSQL)
	if err != nil {
		t.Errorf("Could not parse migration sections: %s", err.Error())
	}
	expectedUp := `INSERT INTO students(first, last) VALUES ('Bobby', 'Tables');`
	if strings.TrimSpace(up) != expectedUp {
		t.Errorf("Parsed up migrations incorrectly. Expected %s, Got: %s", expectedUp, up)
	}

	down, err := extractDown(testSQL)
	if err != nil {
		t.Errorf("Could not parse migration sections: %s", err.Error())
	}
	expectedDown := `DROP TABLE students;`
	if strings.TrimSpace(down) != expectedDown {
		t.Errorf("Parsed down migrations incorrectly. Expected: %s, Got: %s", expectedDown, down)
	}
}
