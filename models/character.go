package models

import (
	"encoding/json"
	"os"
	"path/filepath"
)

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
	SaveFile           string       `json:"-"`
}

func (c *Character) SaveToFile() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(c.SaveFile)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(c.SaveFile, data, 0644)
}

func LoadCharacterFromFile(filename string) (*Character, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var c Character
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	c.SaveFile = filename
	return &c, nil
}

type Feature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
