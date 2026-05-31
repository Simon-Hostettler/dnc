package styles

import "hostettler.dev/dnc/models"

type EnumMapping struct {
	Value int
	Label string
}

func ToSymbol(p models.Proficiency) string {
	for _, m := range ProficiencySymbols {
		if int(p) == m.Value {
			return m.Label
		}
	}
	return ""
}

var ProficiencySymbols []EnumMapping = []EnumMapping{
	{Value: int(models.NoProficiency), Label: "○"},
	{Value: int(models.Proficient), Label: "◐"},
	{Value: int(models.Expertise), Label: "●"},
}

var BinarySymbols []EnumMapping = []EnumMapping{
	{Value: 0, Label: "□"},
	{Value: 1, Label: "■"},
}

var SpellSourceStrings []EnumMapping = []EnumMapping{
	{Value: int(models.InSpellbook), Label: "In Spellbook"},
	{Value: int(models.Temporary), Label: "Temporary"},
}

var SpellSourceSymbols []EnumMapping = []EnumMapping{
	{Value: int(models.InSpellbook), Label: ""},
	{Value: int(models.Temporary), Label: "⧖"},
}

var IsEquippableSymbols []EnumMapping = []EnumMapping{
	{Value: 0, Label: "Not Equippable"},
	{Value: 1, Label: "Equippable"},
}

var AttunementSymbols []EnumMapping = []EnumMapping{
	{Value: 0, Label: "□□□"},
	{Value: 1, Label: "■□□"},
	{Value: 2, Label: "■■□"},
	{Value: 3, Label: "■■■"},
}

var ExhaustionSymbols []EnumMapping = []EnumMapping{
	{Value: 0, Label: "□□□□□□"},
	{Value: 1, Label: "■□□□□□"},
	{Value: 2, Label: "■■□□□□"},
	{Value: 3, Label: "■■■□□□"},
	{Value: 4, Label: "■■■■□□"},
	{Value: 5, Label: "■■■■■□"},
	{Value: 6, Label: "■■■■■■"},
}

var DeathSaveSymbols []EnumMapping = []EnumMapping{
	{Value: 0, Label: "○○○"},
	{Value: 1, Label: "●○○"},
	{Value: 2, Label: "●●○"},
	{Value: 3, Label: "●●●"},
}
