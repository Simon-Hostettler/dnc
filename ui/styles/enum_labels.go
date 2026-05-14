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

var PreparedSymbols []EnumMapping = []EnumMapping{
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

var EquippedSymbols []EnumMapping = []EnumMapping{
	{Value: int(models.NonEquippable), Label: "Not Equippable"},
	{Value: int(models.NotEquipped), Label: "Not Equipped"},
	{Value: int(models.Equipped), Label: "Equipped"},
}

var AttunementSymbols []EnumMapping = []EnumMapping{
	{Value: 0, Label: "□□□"},
	{Value: 1, Label: "■□□"},
	{Value: 2, Label: "■■□"},
	{Value: 3, Label: "■■■"},
}

var DeathSaveSymbols []EnumMapping = []EnumMapping{
	{Value: 0, Label: "○○○"},
	{Value: 1, Label: "●○○"},
	{Value: 2, Label: "●●○"},
	{Value: 3, Label: "●●●"},
}
