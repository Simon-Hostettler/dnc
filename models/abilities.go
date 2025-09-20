package models

import (
	"errors"

	structs "github.com/fatih/structs"
)

type Abilities struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

func (a Abilities) Get(ability string) (int, error) {
	m := structs.Map(a)
	if val, ok := m[ability]; ok {
		return val.(int), nil
	} else {
		return -1, errors.New("undefined ability")
	}
}

func ToModifier(score int) int {
	return (score - 10) / 2
}

type ProficiencyLevel int

const (
	NoProficiency ProficiencyLevel = iota
	Proficient
	Expertise
)

type Skill struct {
	Name           string           `json:"name"`
	Ability        string           `json:"ability"`
	Proficiency    ProficiencyLevel `json:"proficiency"`
	CustomModifier int              `json:"custom_modifier"`
}

func (s Skill) ToModifier(a Abilities, profBonus int) int {
	base, err := a.Get(s.Ability)
	if err != nil {
		return 0
	}
	return ToModifier(base) + int(s.Proficiency)*profBonus + s.CustomModifier
}

func NewSkills() []Skill {
	return []Skill{
		{Name: "Athletics", Ability: "Strength", Proficiency: NoProficiency},
		{Name: "Acrobatics", Ability: "Dexterity", Proficiency: NoProficiency},
		{Name: "Sleight of Hand", Ability: "Dexterity", Proficiency: NoProficiency},
		{Name: "Stealth", Ability: "Dexterity", Proficiency: NoProficiency},
		{Name: "Arcana", Ability: "Intelligence", Proficiency: NoProficiency},
		{Name: "History", Ability: "Intelligence", Proficiency: NoProficiency},
		{Name: "Investigation", Ability: "Intelligence", Proficiency: NoProficiency},
		{Name: "Nature", Ability: "Intelligence", Proficiency: NoProficiency},
		{Name: "Religion", Ability: "Intelligence", Proficiency: NoProficiency},
		{Name: "Animal Handling", Ability: "Wisdom", Proficiency: NoProficiency},
		{Name: "Insight", Ability: "Wisdom", Proficiency: NoProficiency},
		{Name: "Medicine", Ability: "Wisdom", Proficiency: NoProficiency},
		{Name: "Perception", Ability: "Wisdom", Proficiency: NoProficiency},
		{Name: "Survival", Ability: "Wisdom", Proficiency: NoProficiency},
		{Name: "Deception", Ability: "Charisma", Proficiency: NoProficiency},
		{Name: "Intimidation", Ability: "Charisma", Proficiency: NoProficiency},
		{Name: "Performance", Ability: "Charisma", Proficiency: NoProficiency},
		{Name: "Persuasion", Ability: "Charisma", Proficiency: NoProficiency},
	}
}

type SavingThrow struct {
	Ability     string           `json:"ability"`
	Proficiency ProficiencyLevel `json:"proficiency"`
}

func (s SavingThrow) ToModifier(a Abilities, profBonus int) int {
	base, err := a.Get(s.Ability)
	if err != nil {
		return 0
	}
	return ToModifier(base) + int(s.Proficiency)*profBonus
}

type SavingThrows struct {
	Strength     SavingThrow `json:"strength"`
	Dexterity    SavingThrow `json:"dexterity"`
	Constitution SavingThrow `json:"constitution"`
	Intelligence SavingThrow `json:"intelligence"`
	Wisdom       SavingThrow `json:"wisdom"`
	Charisma     SavingThrow `json:"charisma"`
}

func NewSavingThrows() []SavingThrow {
	return []SavingThrow{
		{Ability: "Strength"},
		{Ability: "Dexterity"},
		{Ability: "Constitution"},
		{Ability: "Intelligence"},
		{Ability: "Wisdom"},
		{Ability: "Charisma"},
	}
}
