package models

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Character struct {
	Name               string        `json:"name"`
	Race               string        `json:"race"`
	ClassLevels        string        `json:"class_levels"`
	Background         string        `json:"background"`
	Alignment          string        `json:"alignment"`
	Inspiration        bool          `json:"inspiration"`
	Abilities          Abilities     `json:"abilities"`
	ProficiencyBonus   int           `json:"proficiency_bonus"`
	Skills             []Skill       `json:"skills"`
	SavingThrows       []SavingThrow `json:"saving_throws"`
	ArmorClass         int           `json:"armor_class"`
	Initiative         int           `json:"initiative"`
	Speed              int           `json:"speed"`
	MaxHitPoints       int           `json:"max_hit_points"`
	CurrentHitPoints   int           `json:"current_hit_points"`
	TemporaryHitPoints int           `json:"temporary_hit_points"`
	HitDice            string        `json:"hit_dice"`
	UsedHitDice        string        `json:"used_hit_dice"`
	DeathSaves         DeathSaves    `json:"death_saves"`
	Actions            string        `json:"actions"`
	BonusActions       string        `json:"bonus_actions"`
	Attacks            []Attack      `json:"attacks"`
	Equipment          []Item        `json:"equipment"`
	Features           []Feature     `json:"features"`
	Traits             []string      `json:"traits"`
	Spells             Spellcasting  `json:"spells"`
	Languages          []string      `json:"languages"`
	PersonalityTraits  string        `json:"personality_traits"`
	Ideals             string        `json:"ideals"`
	Bonds              string        `json:"bonds"`
	Flaws              string        `json:"flaws"`
	SaveFile           string        `json:"-"`
}

func NewCharacter(name string) (Character, error) {
	c := Character{
		Name:         name,
		Skills:       NewSkills(),
		SavingThrows: NewSavingThrows(),
		Spells:       NewSpellcasting(),
	}
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return c, err
	}
	characterFile := filepath.Join(
		cfgDir,
		"dnc",
		"characters",
		strings.ToLower(name)+".json",
	)
	if _, err := os.Stat(characterFile); errors.Is(err, os.ErrNotExist) {
		c.SaveFile = characterFile
		return c, nil
	} else {
		return c, errors.New("character already exists")
	}
}

func (c *Character) SaveToFile() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(c.SaveFile)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(c.SaveFile, data, 0o644)
}

func (c *Character) GetSpellsByLevel(l int) []*Spell {
	spells := []*Spell{}
	for i := range c.Spells.SpellsKnown {
		if c.Spells.SpellsKnown[i].Level == l {
			spells = append(spells, &c.Spells.SpellsKnown[i])
		}
	}
	return spells
}

func (c *Character) AddEmptyAttack() {
	c.Attacks = append(c.Attacks, Attack{})
}

func LoadCharacterByName(name string) (*Character, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	characterFile := filepath.Join(
		cfgDir,
		"dnc",
		"characters",
		strings.ToLower(name)+".json",
	)

	data, err := os.ReadFile(characterFile)
	if err != nil {
		return nil, err
	}
	var c Character
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	c.SaveFile = characterFile
	return &c, nil
}

type Feature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
