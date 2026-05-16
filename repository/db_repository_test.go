package repository

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"hostettler.dev/dnc/db"
	"hostettler.dev/dnc/models"
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
	if diff := cmp.Diff(testChar, *loaded, diffIgnoringTimestampsOption(), cmpopts.IgnoreUnexported(CharacterAggregate{})); diff != "" {
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

// childTablesUnderTest mirrors the cascade list in DBCharacterRepository.Delete.
var childTablesUnderTest = []string{
	"wallet", "abilities", "saving_throws",
	"item", "spell", "attacks", "character_skill", "features", "notes",
}

// newTestRepo bootstraps a migrated temp DB and registers its teardown so a
// t.Fatalf mid-test still tears down.
func newTestRepo(t *testing.T) (*DBCharacterRepository, *sqlx.DB) {
	t.Helper()
	dbPath := db.TestDBPath()
	handle, err := db.TestDBInstance(dbPath)
	if err != nil {
		t.Fatalf("Could not create test DB: %s", err.Error())
	}
	if err := db.MigrateUp(handle); err != nil {
		t.Fatalf("Migration to current version failed: %s", err.Error())
	}
	t.Cleanup(func() {
		if err := db.DestroyTestDB(handle, dbPath); err != nil {
			t.Fatalf("Could not destroy test DB: %s", err.Error())
		}
	})
	return NewDBCharacterRepository(handle), handle
}

func countRows(t *testing.T, h *sqlx.DB, table string, charID uuid.UUID) int {
	t.Helper()
	var n int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE character_id = ?", table)
	if err := h.Get(&n, query, charID); err != nil {
		t.Fatalf("Could not count rows in %s: %s", table, err.Error())
	}
	return n
}

func TestDeleteCascadesAllChildTables(t *testing.T) {
	repo, handle := newTestRepo(t)
	ctx := context.Background()

	id, err := repo.CreateEmpty(ctx, "Bobby")
	if err != nil {
		t.Fatalf("Could not create character: %s", err.Error())
	}
	testChar := TestCharacter(id)
	if err := repo.Update(ctx, &testChar); err != nil {
		t.Fatalf("Could not populate character: %s", err.Error())
	}

	// Guard against a false green: confirm the fixture actually populated
	// every child table before asserting the cascade emptied them.
	for _, table := range childTablesUnderTest {
		if n := countRows(t, handle, table, id); n == 0 {
			t.Fatalf("fixture left %s empty; cascade assertion would be meaningless", table)
		}
	}

	if err := repo.Delete(ctx, id); err != nil {
		t.Fatalf("Could not delete character: %s", err.Error())
	}

	for _, table := range childTablesUnderTest {
		if n := countRows(t, handle, table, id); n != 0 {
			t.Errorf("Delete left %d orphaned rows in %s", n, table)
		}
	}
	var charCount int
	if err := handle.Get(&charCount, `SELECT COUNT(*) FROM character WHERE id = ?`, id); err != nil {
		t.Fatalf("Could not count character rows: %s", err.Error())
	}
	if charCount != 0 {
		t.Errorf("Delete left the character row, count = %d", charCount)
	}
}

func TestDeleteIsScopedToOneCharacter(t *testing.T) {
	repo, handle := newTestRepo(t)
	ctx := context.Background()

	idA, err := repo.CreateEmpty(ctx, "Alice")
	if err != nil {
		t.Fatalf("Could not create character A: %s", err.Error())
	}
	idB, err := repo.CreateEmpty(ctx, "Bob")
	if err != nil {
		t.Fatalf("Could not create character B: %s", err.Error())
	}
	charA := TestCharacter(idA)
	charB := TestCharacter(idB)
	if err := repo.Update(ctx, &charA); err != nil {
		t.Fatalf("Could not populate character A: %s", err.Error())
	}
	if err := repo.Update(ctx, &charB); err != nil {
		t.Fatalf("Could not populate character B: %s", err.Error())
	}

	if err := repo.Delete(ctx, idA); err != nil {
		t.Fatalf("Could not delete character A: %s", err.Error())
	}

	for _, table := range childTablesUnderTest {
		if n := countRows(t, handle, table, idB); n == 0 {
			t.Errorf("Delete of A wiped %s rows for unrelated character B", table)
		}
	}
	if _, err := repo.GetByID(ctx, idB); err != nil {
		t.Errorf("Character B no longer loads after deleting A: %s", err.Error())
	}
}

func TestUpdateReplacesOneToManyAfterMutation(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	id, err := repo.CreateEmpty(ctx, "Bobby")
	if err != nil {
		t.Fatalf("Could not create character: %s", err.Error())
	}
	testChar := TestCharacter(id)
	if err := repo.Update(ctx, &testChar); err != nil {
		t.Fatalf("Could not populate character: %s", err.Error())
	}
	loaded, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("Could not load character: %s", err.Error())
	}

	// Mutate a 1:N collection (drop one note, add another), a 1:1 table, a
	// single-row child, and a scalar on the character row.
	loaded.Notes = append(loaded.Notes[1:], models.NoteTO{
		ID:          uuid.New(),
		CharacterID: id,
		Title:       "Added After Mutation",
		Note:        "fresh note",
	})
	loaded.Items[0].Quantity = 99
	loaded.Wallet.Gold = 999
	loaded.Character.CurrHitPoints = 7

	if err := repo.Update(ctx, loaded); err != nil {
		t.Fatalf("Could not update character: %s", err.Error())
	}
	reloaded, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("Could not reload character: %s", err.Error())
	}

	// GetByID returns notes ordered by title ASC; match that before diffing.
	sort.Slice(loaded.Notes, func(i, j int) bool { return loaded.Notes[i].Title < loaded.Notes[j].Title })

	if diff := cmp.Diff(*loaded, *reloaded, diffIgnoringTimestampsOption(), cmpopts.IgnoreUnexported(CharacterAggregate{})); diff != "" {
		t.Errorf("Mismatch between mutated and reloaded character:\n%s", diff)
	}
}

// a save that only mutates the character row must not touch
// created_at / updated_at on any child table.
func TestUpdateCharacterOnlyPreservesChildTimestamps(t *testing.T) {
	repo, handle := newTestRepo(t)
	ctx := context.Background()

	id, err := repo.CreateEmpty(ctx, "Bobby")
	if err != nil {
		t.Fatalf("Could not create character: %s", err.Error())
	}
	testChar := TestCharacter(id)
	if err := repo.Update(ctx, &testChar); err != nil {
		t.Fatalf("Could not populate character: %s", err.Error())
	}

	// Reload to capture the canonical baseline timestamps assigned by the DB.
	loaded, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("Could not load character: %s", err.Error())
	}
	before := collectChildTimestamps(t, handle, id)
	if len(before) == 0 {
		t.Fatal("fixture left no child rows; assertion would be meaningless")
	}

	// Touch only the character row.
	loaded.Character.CurrHitPoints = 3
	if err := repo.Update(ctx, loaded); err != nil {
		t.Fatalf("Could not update character: %s", err.Error())
	}

	after := collectChildTimestamps(t, handle, id)
	if diff := cmp.Diff(before, after); diff != "" {
		t.Errorf("child timestamps moved after a character-only update:\n%s", diff)
	}
}

// mutating a single note must not bump the updated_at
// of rows in unrelated child tables.
func TestUpdateOnlyRewritesDirtySection(t *testing.T) {
	repo, handle := newTestRepo(t)
	ctx := context.Background()

	id, err := repo.CreateEmpty(ctx, "Bobby")
	if err != nil {
		t.Fatalf("Could not create character: %s", err.Error())
	}
	testChar := TestCharacter(id)
	if err := repo.Update(ctx, &testChar); err != nil {
		t.Fatalf("Could not populate character: %s", err.Error())
	}
	loaded, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("Could not load character: %s", err.Error())
	}

	before := collectChildTimestamps(t, handle, id)
	if len(loaded.Notes) == 0 {
		t.Fatal("fixture has no notes; assertion would be meaningless")
	}

	loaded.Notes[0].Title = "Edited title"
	if err := repo.Update(ctx, loaded); err != nil {
		t.Fatalf("Could not update character: %s", err.Error())
	}
	after := collectChildTimestamps(t, handle, id)

	for key, beforeTs := range before {
		afterTs, ok := after[key]
		if !ok {
			t.Errorf("row %s disappeared after update", key)
			continue
		}
		if key[:len("notes:")] == "notes:" {
			// Notes section rewritten — created_at must be preserved for the
			// unedited rows. The whole-section rewrite legitimately bumps
			// updated_at across the section, which is the accepted tradeoff.
			if afterTs.Created != beforeTs.Created {
				t.Errorf("note %s lost its original created_at", key)
			}
			continue
		}
		if afterTs != beforeTs {
			t.Errorf("untouched section row %s timestamps moved: before=%+v after=%+v", key, beforeTs, afterTs)
		}
	}
}

type rowTimestamps struct {
	Created time.Time
	Updated time.Time
}

// map of "<table>:<id>" -> (created_at, updated_at) for every child row
func collectChildTimestamps(t *testing.T, h *sqlx.DB, charID uuid.UUID) map[string]rowTimestamps {
	t.Helper()
	out := map[string]rowTimestamps{}
	// 1:N tables keyed by their own id.
	perRow := []struct{ table, idCol string }{
		{"item", "id"},
		{"spell", "id"},
		{"attacks", "id"},
		{"features", "id"},
		{"notes", "id"},
		{"character_skill", "id"},
	}
	for _, p := range perRow {
		rows, err := h.Queryx(fmt.Sprintf("SELECT %s::TEXT, created_at, updated_at FROM %s WHERE character_id=?", p.idCol, p.table), charID)
		if err != nil {
			t.Fatalf("query %s: %s", p.table, err.Error())
		}
		for rows.Next() {
			var idStr string
			var c, u time.Time
			if err := rows.Scan(&idStr, &c, &u); err != nil {
				rows.Close()
				t.Fatalf("scan %s: %s", p.table, err.Error())
			}
			out[p.table+":"+idStr] = rowTimestamps{c, u}
		}
		rows.Close()
	}
	// 1:1 tables keyed by character_id.
	for _, table := range []string{"abilities", "saving_throws", "wallet"} {
		var c, u time.Time
		err := h.QueryRow(fmt.Sprintf("SELECT created_at, updated_at FROM %s WHERE character_id=?", table), charID).Scan(&c, &u)
		if err != nil {
			t.Fatalf("query %s: %s", table, err.Error())
		}
		out[table+":"+charID.String()] = rowTimestamps{c, u}
	}
	return out
}
