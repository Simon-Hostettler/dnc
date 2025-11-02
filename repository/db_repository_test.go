package repository

import (
	"context"
	"testing"

	"hostettler.dev/dnc/db"
)

func TestEmptyOperations(t *testing.T) {
	dbPath := db.TestDBPath()
	handle, err := db.TestDBInstance(dbPath)
	if err != nil {
		t.Fatalf("Could not create test DB: %s", err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := db.MigrateUp(handle); err != nil {
		t.Fatalf("Migration to current version failed: %s", err.Error())
	}
	repo := NewDBCharacterRepository(handle)
	id, err := repo.CreateEmpty(ctx, "Bobby")
	if err != nil {
		t.Fatalf("Could not create a new character: %s", err.Error())
	}
	c, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("Could not retrieve the character: %s", err.Error())
	}
	err = repo.Update(ctx, c)
	if err != nil {
		t.Errorf("Could not update the character without changes: %s", err.Error())
	}
	err = repo.Delete(ctx, id)
	if err != nil {
		t.Errorf("Could not delete the character: %s", err.Error())
	}
	cancel() // just to be sure
	if err := db.DestroyTestDB(handle, dbPath); err != nil {
		t.Fatalf("Could not destroy test DB: %s", err.Error())
	}
}
