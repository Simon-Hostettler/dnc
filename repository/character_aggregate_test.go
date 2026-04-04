package repository

import (
	"testing"

	"github.com/google/uuid"
	"hostettler.dev/dnc/models"
)

func newTestAggregate() *CharacterAggregate {
	return &CharacterAggregate{
		Character: &models.CharacterTO{},
	}
}

func TestAddEmptyItem(t *testing.T) {
	agg := newTestAggregate()
	id1 := agg.AddEmptyItem()
	id2 := agg.AddEmptyItem()

	if len(agg.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(agg.Items))
	}
	if id1 == id2 {
		t.Error("expected unique IDs")
	}
	if agg.Items[0].ID != id1 || agg.Items[1].ID != id2 {
		t.Error("IDs don't match returned values")
	}
}

func TestAddEmptySpellSetsLevel(t *testing.T) {
	agg := newTestAggregate()
	agg.AddEmptySpell(3)
	agg.AddEmptySpell(0)

	if len(agg.Spells) != 2 {
		t.Fatalf("expected 2 spells, got %d", len(agg.Spells))
	}
	if agg.Spells[0].Level != 3 {
		t.Errorf("first spell level = %d, want 3", agg.Spells[0].Level)
	}
	if agg.Spells[1].Level != 0 {
		t.Errorf("second spell level = %d, want 0", agg.Spells[1].Level)
	}
}

func TestDeleteItem(t *testing.T) {
	agg := newTestAggregate()
	id1 := agg.AddEmptyItem()
	id2 := agg.AddEmptyItem()
	id3 := agg.AddEmptyItem()

	agg.DeleteItem(id2)

	if len(agg.Items) != 2 {
		t.Fatalf("expected 2 items after delete, got %d", len(agg.Items))
	}
	for _, item := range agg.Items {
		if item.ID == id2 {
			t.Error("deleted item still present")
		}
	}
	if agg.Items[0].ID != id1 || agg.Items[1].ID != id3 {
		t.Error("remaining items are wrong")
	}
}

func TestDeleteNonExistentIsNoOp(t *testing.T) {
	agg := newTestAggregate()
	agg.AddEmptyItem()
	agg.DeleteItem(uuid.New()) // non-existent ID

	if len(agg.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(agg.Items))
	}
}

func TestDeleteFromSingleElementSlice(t *testing.T) {
	agg := newTestAggregate()
	id := agg.AddEmptyAttack()
	agg.DeleteAttack(id)

	if len(agg.Attacks) != 0 {
		t.Errorf("expected 0 attacks, got %d", len(agg.Attacks))
	}
}

func TestDeleteSpell(t *testing.T) {
	agg := newTestAggregate()
	agg.AddEmptySpell(1)
	id := agg.AddEmptySpell(2)
	agg.AddEmptySpell(3)

	agg.DeleteSpell(id)
	if len(agg.Spells) != 2 {
		t.Fatalf("expected 2 spells, got %d", len(agg.Spells))
	}
}

func TestDeleteFeature(t *testing.T) {
	agg := newTestAggregate()
	id := agg.AddEmptyFeature()
	agg.DeleteFeature(id)
	if len(agg.Features) != 0 {
		t.Errorf("expected 0 features, got %d", len(agg.Features))
	}
}

func TestDeleteNote(t *testing.T) {
	agg := newTestAggregate()
	id := agg.AddEmptyNote()
	agg.DeleteNote(id)
	if len(agg.Notes) != 0 {
		t.Errorf("expected 0 notes, got %d", len(agg.Notes))
	}
}

func TestGetSpellsByLevel(t *testing.T) {
	agg := newTestAggregate()
	agg.AddEmptySpell(0) // cantrip
	agg.AddEmptySpell(1)
	agg.AddEmptySpell(1)
	agg.AddEmptySpell(3)
	agg.AddEmptySpell(1)

	level1 := agg.GetSpellsByLevel(1)
	if len(level1) != 3 {
		t.Errorf("expected 3 level-1 spells, got %d", len(level1))
	}

	cantrips := agg.GetSpellsByLevel(0)
	if len(cantrips) != 1 {
		t.Errorf("expected 1 cantrip, got %d", len(cantrips))
	}

	level9 := agg.GetSpellsByLevel(9)
	if len(level9) != 0 {
		t.Errorf("expected 0 level-9 spells, got %d", len(level9))
	}
}

func TestGetSpellsByLevelReturnsMutablePointers(t *testing.T) {
	agg := newTestAggregate()
	agg.AddEmptySpell(2)

	spells := agg.GetSpellsByLevel(2)
	spells[0].Name = "Fireball"

	if agg.Spells[0].Name != "Fireball" {
		t.Error("expected mutation through pointer to affect aggregate")
	}
}
