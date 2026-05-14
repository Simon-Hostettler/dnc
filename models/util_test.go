package models

import "testing"

func TestToModifier(t *testing.T) {
	tests := []struct {
		name      string
		score     int
		prof      Proficiency
		profBonus int
		want      int
	}{
		// (score-10)>>1 is an arithmetic shift, not integer division by 2:
		// these negative/odd cases are the ones a naive refactor would break.
		{"negative even base", 8, NoProficiency, 2, -1},
		{"negative odd base", 9, NoProficiency, 2, -1},
		{"positive odd base truncates", 11, NoProficiency, 2, 0},
		{"score 10 boundary", 10, NoProficiency, 2, 0},
		{"plain modifier", 14, NoProficiency, 3, 2},
		{"proficient adds bonus once", 14, Proficient, 3, 5},
		{"expertise adds bonus twice", 14, Expertise, 3, 8},
		{"high score expertise", 20, Expertise, 4, 13},
		{"negative base with expertise", 8, Expertise, 2, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToModifier(tt.score, tt.prof, tt.profBonus); got != tt.want {
				t.Errorf("ToModifier(%d, %d, %d) = %d, want %d", tt.score, tt.prof, tt.profBonus, got, tt.want)
			}
		})
	}
}

func TestToScoreByName(t *testing.T) {
	// Distinct values per ability so a wrong-field lookup is caught.
	abilities := AbilitiesTO{
		Strength:     15,
		Dexterity:    14,
		Constitution: 13,
		Intelligence: 12,
		Wisdom:       10,
		Charisma:     8,
	}
	tests := []struct {
		name    string
		ability string
		want    int
	}{
		{"strength", "strength", 15},
		{"dexterity", "dexterity", 14},
		{"constitution", "constitution", 13},
		{"intelligence", "intelligence", 12},
		{"wisdom", "wisdom", 10},
		{"charisma", "charisma", 8},
		{"title case", "Strength", 15},
		{"upper case", "STRENGTH", 15},
		{"mixed case", "DeXtErItY", 14},
		{"unknown ability", "luck", 0},
		{"empty string", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := abilities.ToScoreByName(tt.ability); got != tt.want {
				t.Errorf("ToScoreByName(%q) = %d, want %d", tt.ability, got, tt.want)
			}
		})
	}
}
