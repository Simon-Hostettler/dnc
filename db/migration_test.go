package db

import (
	"strings"
	"testing"
)

func TestEmptyMigrations(t *testing.T) {
	dbPath := TestDBPath()
	handle, err := TestDBInstance(dbPath)
	if err != nil {
		t.Fatalf("Could not create test DB: %s", err.Error())
	}
	if err := MigrateUp(handle); err != nil {
		t.Errorf("Migration to current version failed: %s", err.Error())
	}
	if err := MigrateDown(handle); err != nil {
		t.Errorf("Migrating down to initial DB failed: %s", err.Error())
	}
	if err := DestroyTestDB(handle, dbPath); err != nil {
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
