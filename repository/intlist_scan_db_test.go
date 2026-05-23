package repository

import (
	"testing"

	"hostettler.dev/dnc/db"
)

// Locks assumption that models.IntList.Scan relies on:
// DuckDB INTEGER[] column arrives as []any with int32 element
// Rewrite IntList.Scanif driver ever changes representation
func TestDuckDBIntegerListScanType(t *testing.T) {
	dbPath := db.TestDBPath()
	handle, err := db.TestDBInstance(dbPath)
	if err != nil {
		t.Fatalf("Could not create test DB: %s", err.Error())
	}
	defer func() {
		if err := db.DestroyTestDB(handle, dbPath); err != nil {
			t.Fatalf("Could not destroy test DB: %s", err.Error())
		}
	}()

	var dest any
	if err := handle.QueryRow(`SELECT [1, 2, 3]::INTEGER[]`).Scan(&dest); err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	slice, ok := dest.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", dest)
	}
	if len(slice) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(slice))
	}
	for i, e := range slice {
		if _, ok := e.(int32); !ok {
			t.Errorf("element %d: expected int32, got %T", i, e)
		}
	}
}
