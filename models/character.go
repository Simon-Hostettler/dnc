package models

type Character struct {
	Name               string       `json:"name"`
	Race               string       `json:"race"`
	Class              string       `json:"class"`
	Level              int          `json:"level"`
	Background         string       `json:"background"`
	Alignment          string       `json:"alignment"`
	Inspiration        bool         `json:"inspiration"`
	Abilities          Abilities    `json:"abilities"`
	ProficiencyBonus   int          `json:"proficiency_bonus"`
	Skills             Skills       `json:"skills"`
	ArmorClass         int          `json:"armor_class"`
	Initiative         int          `json:"initiative"`
	Speed              int          `json:"speed"`
	MaxHitPoints       int          `json:"max_hit_points"`
	CurrentHitPoints   int          `json:"current_hit_points"`
	TemporaryHitPoints int          `json:"temporary_hit_points"`
	HitDice            string       `json:"hit_dice"`
	DeathSaves         DeathSaves   `json:"death_saves"`
	Attacks            []Attack     `json:"attacks"`
	Equipment          []Item       `json:"equipment"`
	Features           []Feature    `json:"features"`
	Traits             []string     `json:"traits"`
	Spells             Spellcasting `json:"spells"`
	Languages          []string     `json:"languages"`
	PersonalityTraits  string       `json:"personality_traits"`
	Ideals             string       `json:"ideals"`
	Bonds              string       `json:"bonds"`
	Flaws              string       `json:"flaws"`
}

type Feature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
