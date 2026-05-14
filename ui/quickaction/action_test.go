package quickaction

import (
	"strings"
	"testing"

	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/repository"
)

// charAgg builds a minimal aggregate carrying only the fields the actions touch.
// Keep slots and used the same length: CastAction indexes SpellSlotsUsed[level]
// after only bounds-checking SpellSlots.
func charAgg(curr, max int, slots, used []int) *repository.CharacterAggregate {
	return &repository.CharacterAggregate{
		Character: &models.CharacterTO{
			CurrHitPoints:  curr,
			MaxHitPoints:   max,
			SpellSlots:     slots,
			SpellSlotsUsed: used,
		},
	}
}

func assertWriteBack(t *testing.T, r ActionResult) {
	t.Helper()
	if r.ErrMsg != "" {
		t.Errorf("expected no error, got ErrMsg = %q", r.ErrMsg)
	}
	if r.Cmd == nil {
		t.Fatal("expected a write-back Cmd, got nil")
	}
	if _, ok := r.Cmd().(command.WriteBackRequestMsg); !ok {
		t.Errorf("expected WriteBackRequestMsg, got %T", r.Cmd())
	}
}

func assertErr(t *testing.T, r ActionResult, wantSubstr string) {
	t.Helper()
	if r.Cmd != nil {
		t.Error("expected no Cmd on error, got non-nil")
	}
	if !strings.Contains(r.ErrMsg, wantSubstr) {
		t.Errorf("expected ErrMsg containing %q, got %q", wantSubstr, r.ErrMsg)
	}
}

func TestLongRestAction(t *testing.T) {
	agg := charAgg(3, 50, []int{0, 4, 2}, []int{0, 4, 2})
	agg.Character.DeathSaveSuccesses = 2
	agg.Character.DeathSaveFailures = 1

	res := LongRestAction{}.Execute(agg, "")

	c := agg.Character
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
	assertWriteBack(t, res)
}

func TestCastAction(t *testing.T) {
	tests := []struct {
		name     string
		slots    []int
		used     []int
		args     string
		wantErr  string // empty => expect write-back success
		checkIdx int    // index of SpellSlotsUsed to inspect afterwards
		wantUsed int    // expected SpellSlotsUsed[checkIdx]
	}{
		{"valid cast increments slot", []int{0, 0, 0, 2}, []int{0, 0, 0, 0}, "3", "", 3, 1},
		{"non-numeric arg", []int{0, 2}, []int{0, 0}, "abc", "usage", 1, 0},
		{"empty arg", []int{0, 2}, []int{0, 0}, "", "usage", 1, 0},
		{"level below range", []int{0, 2}, []int{0, 0}, "0", "usage", 1, 0},
		{"level above range", make([]int, 10), make([]int, 10), "10", "usage", 1, 0},
		{"no slots at level", []int{0, 0, 0, 0, 0, 0}, []int{0, 0, 0, 0, 0, 0}, "5", "no spell slots at level 5", 5, 0},
		{"level beyond slice", []int{0, 1, 1}, []int{0, 0, 0}, "5", "no spell slots", 2, 0},
		{"all slots used", []int{0, 0, 2}, []int{0, 0, 2}, "2", "no available slots at level 2", 2, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := charAgg(10, 10, tt.slots, tt.used)
			res := CastAction{}.Execute(agg, tt.args)
			if tt.wantErr == "" {
				assertWriteBack(t, res)
			} else {
				assertErr(t, res, tt.wantErr)
			}
			// Error paths must not have mutated the slot counters.
			if got := agg.Character.SpellSlotsUsed[tt.checkIdx]; got != tt.wantUsed {
				t.Errorf("SpellSlotsUsed[%d] = %d, want %d", tt.checkIdx, got, tt.wantUsed)
			}
		})
	}
}

func TestHealAction(t *testing.T) {
	tests := []struct {
		name     string
		curr     int
		max      int
		args     string
		wantErr  string
		wantCurr int
	}{
		{"normal heal", 10, 100, "20", "", 30},
		{"clamps to max", 90, 100, "50", "", 100},
		{"negative amount", 50, 100, "-5", "usage", 50},
		{"non-numeric amount", 50, 100, "x", "usage", 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := charAgg(tt.curr, tt.max, nil, nil)
			res := HealAction{}.Execute(agg, tt.args)
			if tt.wantErr == "" {
				assertWriteBack(t, res)
			} else {
				assertErr(t, res, tt.wantErr)
			}
			if agg.Character.CurrHitPoints != tt.wantCurr {
				t.Errorf("CurrHitPoints = %d, want %d", agg.Character.CurrHitPoints, tt.wantCurr)
			}
		})
	}
}

func TestDmgAction(t *testing.T) {
	tests := []struct {
		name     string
		curr     int
		max      int
		args     string
		wantErr  string
		wantCurr int
	}{
		{"normal damage", 50, 100, "20", "", 30},
		{"clamps to zero", 10, 100, "50", "", 0},
		{"negative amount", 50, 100, "-3", "usage", 50},
		{"non-numeric amount", 50, 100, "y", "usage", 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := charAgg(tt.curr, tt.max, nil, nil)
			res := DmgAction{}.Execute(agg, tt.args)
			if tt.wantErr == "" {
				assertWriteBack(t, res)
			} else {
				assertErr(t, res, tt.wantErr)
			}
			if agg.Character.CurrHitPoints != tt.wantCurr {
				t.Errorf("CurrHitPoints = %d, want %d", agg.Character.CurrHitPoints, tt.wantCurr)
			}
		})
	}
}

// The calculator actions delegate their math to the dicestats package; only
// the DnC-owned arg validation is exercised here.
func TestCalculatorActionsRejectEmptyArgs(t *testing.T) {
	actions := []struct {
		name   string
		action Action
	}{
		{"prob", ProbAction{}},
		{"ev", EvAction{}},
		{"dist", DistAction{}},
	}
	args := []struct{ label, value string }{
		{"empty", ""},
		{"whitespace", "   "},
	}
	for _, a := range actions {
		for _, arg := range args {
			t.Run(a.name+"_"+arg.label, func(t *testing.T) {
				res := a.action.Execute(nil, arg.value)
				assertErr(t, res, "usage")
				if res.Result != "" {
					t.Errorf("expected empty Result, got %q", res.Result)
				}
			})
		}
	}
}
