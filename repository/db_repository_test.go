package repository

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func TestAllValuesPersist(t *testing.T) {
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
	testChar := TestCharacter(id)
	err = repo.Update(ctx, &testChar)
	if err != nil {
		t.Fatalf("Could not update the character: %s", err.Error())
	}
	loaded, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("Could not fetch the character: %s", err.Error())
	}
	if diff := cmp.Diff(testChar, *loaded, diffIgnoringTimestampsOption()); diff != "" {
		t.Errorf("Mismatch between stored and loaded values in character:\n%s", diff)
	}
	cancel() // just to be sure
	if err := db.DestroyTestDB(handle, dbPath); err != nil {
		t.Fatalf("Could not destroy test DB: %s", err.Error())
	}
}

func diffIgnoringTimestampsOption() cmp.Option {
	return cmp.FilterPath(func(p cmp.Path) bool {
		sf, ok := p.Last().(cmp.StructField)
		if !ok {
			return false
		}
		name := sf.Name()
		return name == "CreatedAt" || name == "UpdatedAt"
	}, cmp.Ignore())
}
