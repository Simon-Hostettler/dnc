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

type Skills struct {
	Athletics      Skill `json:"athletics"`
	Acrobatics     Skill `json:"acrobatics"`
	SleightOfHand  Skill `json:"sleight_of_hand"`
	Stealth        Skill `json:"stealth"`
	Arcana         Skill `json:"arcana"`
	History        Skill `json:"history"`
	Investigation  Skill `json:"investigation"`
	Nature         Skill `json:"nature"`
	Religion       Skill `json:"religion"`
	AnimalHandling Skill `json:"animal_handling"`
	Insight        Skill `json:"insight"`
	Medicine       Skill `json:"medicine"`
	Perception     Skill `json:"perception"`
	Survival       Skill `json:"survival"`
	Deception      Skill `json:"deception"`
	Intimidation   Skill `json:"intimidation"`
	Performance    Skill `json:"performance"`
	Persuasion     Skill `json:"persuasion"`
}

func NewSkills() Skills {
	return Skills{
		Athletics:      Skill{Name: "Athletics", Ability: "Strength", Proficiency: NoProficiency},
		Acrobatics:     Skill{Name: "Acrobatics", Ability: "Dexterity", Proficiency: NoProficiency},
		SleightOfHand:  Skill{Name: "Sleight of Hand", Ability: "Dexterity", Proficiency: NoProficiency},
		Stealth:        Skill{Name: "Stealth", Ability: "Dexterity", Proficiency: NoProficiency},
		Arcana:         Skill{Name: "Arcana", Ability: "Intelligence", Proficiency: NoProficiency},
		History:        Skill{Name: "History", Ability: "Intelligence", Proficiency: NoProficiency},
		Investigation:  Skill{Name: "Investigation", Ability: "Intelligence", Proficiency: NoProficiency},
		Nature:         Skill{Name: "Nature", Ability: "Intelligence", Proficiency: NoProficiency},
		Religion:       Skill{Name: "Religion", Ability: "Intelligence", Proficiency: NoProficiency},
		AnimalHandling: Skill{Name: "Animal Handling", Ability: "Wisdom", Proficiency: NoProficiency},
		Insight:        Skill{Name: "Insight", Ability: "Wisdom", Proficiency: NoProficiency},
		Medicine:       Skill{Name: "Medicine", Ability: "Wisdom", Proficiency: NoProficiency},
		Perception:     Skill{Name: "Perception", Ability: "Wisdom", Proficiency: NoProficiency},
		Survival:       Skill{Name: "Survival", Ability: "Wisdom", Proficiency: NoProficiency},
		Deception:      Skill{Name: "Deception", Ability: "Charisma", Proficiency: NoProficiency},
		Intimidation:   Skill{Name: "Intimidation", Ability: "Charisma", Proficiency: NoProficiency},
		Performance:    Skill{Name: "Performance", Ability: "Charisma", Proficiency: NoProficiency},
		Persuasion:     Skill{Name: "Persuasion", Ability: "Charisma", Proficiency: NoProficiency},
	}
}
