package models

import (
	"strings"

	"github.com/google/uuid"
)

type CharacterSummary struct {
	ID   uuid.UUID
	Name string
}

type Proficiency int

const (
	NoProficiency Proficiency = iota
	Proficient
	Expertise
)

type SpellSource int

const (
	InSpellbook SpellSource = iota
	Temporary
)

func (c CharacterSkillDetailTO) ToCharacterSkillTO() CharacterSkillTO {
	return CharacterSkillTO{
		ID:             c.ID,
		CharacterID:    c.CharacterID,
		SkillID:        c.SkillID,
		Proficiency:    c.Proficiency,
		CustomModifier: c.CustomModifier,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func ToModifier(score int, prof Proficiency, profBonus int) int {
	return (score-10)>>1 + profBonus*int(prof)
}

func (a AbilitiesTO) ToScoreByName(ability string) int {
	switch strings.ToLower(ability) {
	case "strength":
		return a.Strength
	case "dexterity":
		return a.Dexterity
	case "constitution":
		return a.Constitution
	case "intelligence":
		return a.Intelligence
	case "wisdom":
		return a.Wisdom
	case "charisma":
		return a.Charisma
	}
	return 0
}
