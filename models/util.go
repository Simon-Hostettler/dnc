package models

import (
	"strings"

	"github.com/google/uuid"
)

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

var ProficiencySymbols []EnumMapping = []EnumMapping{
	{Value: int(NoProficiency), Label: "○"},
	{Value: int(Proficient), Label: "◐"},
	{Value: int(Expertise), Label: "●"},
}

func (p Proficiency) ToSymbol() string {
	for _, m := range ProficiencySymbols {
		if int(p) == m.Value {
			return m.Label
		}
	}
	return ""
}

type EnumMapping struct {
	Value int
	Label string
}

var PreparedSymbols []EnumMapping = []EnumMapping{
	{Value: 0, Label: "□"},
	{Value: 1, Label: "■"},
}

var EquippedSymbols []EnumMapping = []EnumMapping{
	{Value: int(NonEquippable), Label: "Not Equippable"},
	{Value: int(NotEquipped), Label: "Not Equipped"},
	{Value: int(Equipped), Label: "Equipped"},
}

var AttunementSymbols []EnumMapping = []EnumMapping{
	{Value: 0, Label: "□□□"},
	{Value: 1, Label: "■□□"},
	{Value: 2, Label: "■■□"},
	{Value: 3, Label: "■■■"},
}

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
	return (score-10)/2 + profBonus*int(prof)
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
