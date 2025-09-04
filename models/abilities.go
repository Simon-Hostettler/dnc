package models

import (
	"fmt"
	"strings"
)

type AbilityScores struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

func (a AbilityScores) Get(ability string) (int, error) {
	switch strings.ToLower(ability) {
	case "strength":
		return a.Strength, nil
	case "dexterity":
		return a.Dexterity, nil
	case "constitution":
		return a.Constitution, nil
	case "intelligence":
		return a.Intelligence, nil
	case "wisdom":
		return a.Wisdom, nil
	case "charisma":
		return a.Charisma, nil
	default:
		return 0, fmt.Errorf("invalid ability: %s", ability)
	}
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
