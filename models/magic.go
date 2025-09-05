package models

type Spellcasting struct {
	SpellcastingAbility string      `json:"spellcasting_ability"`
	SpellSaveDC         int         `json:"spell_save_dc"`
	SpellAttackBonus    int         `json:"spell_attack_bonus"`
	SpellsKnown         []Spell     `json:"spells_known"`
	SpellSlots          map[int]int `json:"spell_slots"` // key = level, value = slots
}

type Spell struct {
	Name        string `json:"name"`
	Level       int    `json:"level"`
	CastingTime string `json:"casting_time"`
	Range       string `json:"range"`
	Duration    string `json:"duration"`
	Components  string `json:"components"`
	Description string `json:"description"`
}
