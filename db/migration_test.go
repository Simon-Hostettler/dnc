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

func TestApplyUpRollsBackOnBadSQL(t *testing.T) {
	dbPath := TestDBPath()
	handle, err := TestDBInstance(dbPath)
	if err != nil {
		t.Fatalf("Could not create test DB: %s", err.Error())
	}
	defer func() {
		if err := DestroyTestDB(handle, dbPath); err != nil {
			t.Fatalf("Could not destroy test DB: %s", err.Error())
		}
	}()
	if err := ensureMigrationTable(handle); err != nil {
		t.Fatalf("Could not create schema_migrations table: %s", err.Error())
	}

	// First statement is valid, second is garbage: execStatements fails midway.
	badSQL := "CREATE TABLE good_tbl (id INTEGER); THIS IS NOT SQL;"
	if err := applyUp(handle, 999, "999_bad.sql", badSQL); err == nil {
		t.Fatal("applyUp with bad SQL returned nil error, expected failure")
	}

	// The valid first statement must have been rolled back too.
	if _, err := handle.Exec(`SELECT 1 FROM good_tbl`); err == nil {
		t.Error("good_tbl exists after failed migration; transaction did not roll back")
	}

	// The version must not have been recorded.
	var count int
	if err := handle.Get(&count, `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, 999); err != nil {
		t.Fatalf("Could not query schema_migrations: %s", err.Error())
	}
	if count != 0 {
		t.Errorf("schema_migrations recorded version 999 despite rollback, count = %d", count)
	}
}

func TestMigrateUpIsIdempotent(t *testing.T) {
	dbPath := TestDBPath()
	handle, err := TestDBInstance(dbPath)
	if err != nil {
		t.Fatalf("Could not create test DB: %s", err.Error())
	}
	defer func() {
		if err := DestroyTestDB(handle, dbPath); err != nil {
			t.Fatalf("Could not destroy test DB: %s", err.Error())
		}
	}()

	if err := MigrateUp(handle); err != nil {
		t.Fatalf("First MigrateUp failed: %s", err.Error())
	}
	var first int
	if err := handle.Get(&first, `SELECT COUNT(*) FROM schema_migrations`); err != nil {
		t.Fatalf("Could not count applied migrations: %s", err.Error())
	}

	// Re-running must be a no-op: already-applied versions are skipped, so no
	// duplicate rows and no error from re-executing CREATE statements.
	if err := MigrateUp(handle); err != nil {
		t.Fatalf("Second MigrateUp failed: %s", err.Error())
	}
	var second int
	if err := handle.Get(&second, `SELECT COUNT(*) FROM schema_migrations`); err != nil {
		t.Fatalf("Could not count applied migrations: %s", err.Error())
	}

	if first != second {
		t.Errorf("MigrateUp not idempotent: %d migrations after first run, %d after second", first, second)
	}

	files, err := listMigrationFiles()
	if err != nil {
		t.Fatalf("Could not list migration files: %s", err.Error())
	}
	if first != len(files) {
		t.Errorf("Expected %d applied migrations, got %d", len(files), first)
	}
}

func TestExtractSectionsRejectMalformed(t *testing.T) {
	t.Run("missing up section", func(t *testing.T) {
		if _, err := extractUp("DROP TABLE students;\n"); err == nil {
			t.Error("extractUp accepted content with no +duckUp marker")
		}
	})
	t.Run("missing down section", func(t *testing.T) {
		if _, err := extractDown("-- +duckUp\nCREATE TABLE students (id INTEGER);\n"); err == nil {
			t.Error("extractDown accepted content with no +duckDown marker")
		}
	})
}
