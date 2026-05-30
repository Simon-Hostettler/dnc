package repository

import (
	"strings"
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
	agg.AddEmptySpell(0)
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

func TestLongRest(t *testing.T) {
	agg := newTestAggregate()
	c := agg.Character
	c.CurrHitPoints = 3
	c.MaxHitPoints = 50
	c.DeathSaveSuccesses = 2
	c.DeathSaveFailures = 1
	c.SpellSlots = []int{0, 4, 2}
	c.SpellSlotsUsed = []int{0, 4, 2}

	agg.LongRest()

	if c.CurrHitPoints != c.MaxHitPoints {
		t.Errorf("CurrHitPoints = %d, want %d", c.CurrHitPoints, c.MaxHitPoints)
	}
	if c.DeathSaveSuccesses != 0 || c.DeathSaveFailures != 0 {
		t.Errorf("death saves not cleared: successes = %d, failures = %d", c.DeathSaveSuccesses, c.DeathSaveFailures)
	}
	for i, used := range c.SpellSlotsUsed {
		if used != 0 {
			t.Errorf("SpellSlotsUsed[%d] = %d, want 0", i, used)
		}
	}
}

func TestHeal(t *testing.T) {
	tests := []struct {
		name     string
		curr     int
		max      int
		amount   int
		wantCurr int
	}{
		{"adds hit points", 10, 100, 20, 30},
		{"clamps to max", 90, 100, 50, 100},
		{"zero is a no-op", 50, 100, 0, 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := newTestAggregate()
			agg.Character.CurrHitPoints = tt.curr
			agg.Character.MaxHitPoints = tt.max

			agg.Heal(tt.amount)

			if agg.Character.CurrHitPoints != tt.wantCurr {
				t.Errorf("CurrHitPoints = %d, want %d", agg.Character.CurrHitPoints, tt.wantCurr)
			}
		})
	}
}

func TestTakeDamage(t *testing.T) {
	tests := []struct {
		name     string
		curr     int
		temp     int
		amount   int
		wantCurr int
		wantTemp int
	}{
		{"reduces current hit points", 50, 0, 20, 30, 0},
		{"clamps current to zero", 10, 0, 50, 0, 0},
		{"absorbed entirely by temp hp", 50, 10, 6, 50, 4},
		{"spills past temp hp onto current", 50, 10, 25, 35, 0},
		{"temp hp exactly absorbs damage", 50, 10, 10, 50, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := newTestAggregate()
			agg.Character.CurrHitPoints = tt.curr
			agg.Character.TempHitPoints = tt.temp

			agg.TakeDamage(tt.amount)

			if agg.Character.CurrHitPoints != tt.wantCurr {
				t.Errorf("CurrHitPoints = %d, want %d", agg.Character.CurrHitPoints, tt.wantCurr)
			}
			if agg.Character.TempHitPoints != tt.wantTemp {
				t.Errorf("TempHitPoints = %d, want %d", agg.Character.TempHitPoints, tt.wantTemp)
			}
		})
	}
}

func TestCastSpell(t *testing.T) {
	tests := []struct {
		name     string
		slots    []int
		used     []int
		level    int
		wantErr  string // empty => expect success
		checkIdx int
		wantUsed int
	}{
		{"consumes an available slot", []int{0, 0, 0, 2}, []int{0, 0, 0, 0}, 3, "", 3, 1},
		{"no slots at level", []int{0, 0, 0, 0, 0, 0}, []int{0, 0, 0, 0, 0, 0}, 5, "no spell slots at level 5", 5, 0},
		{"level beyond slice", []int{0, 1, 1}, []int{0, 0, 0}, 5, "no spell slots", 2, 0},
		{"all slots used", []int{0, 0, 2}, []int{0, 0, 2}, 2, "no available slots at level 2", 2, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := newTestAggregate()
			agg.Character.SpellSlots = tt.slots
			agg.Character.SpellSlotsUsed = tt.used

			err := agg.CastSpell(tt.level)

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want substring %q", err.Error(), tt.wantErr)
				}
			}
			if got := agg.Character.SpellSlotsUsed[tt.checkIdx]; got != tt.wantUsed {
				t.Errorf("SpellSlotsUsed[%d] = %d, want %d", tt.checkIdx, got, tt.wantUsed)
			}
		})
	}
}
