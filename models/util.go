package models

import "github.com/google/uuid"

type CharacterSummary struct {
	ID   uuid.UUID
	Name string
}

type Equippable int

const (
	NonEquippable Equippable = iota
	NotEquipped
	Equipped
)

type Proficiency int

const (
	NoProficiency Proficiency = iota
	Proficient
	Expertise
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

func (c CharacterSkillDetailTO) ToModifier(score int, profMod int) int {
	return (score-10)/2 + profMod*c.Proficiency
}
