package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/structs"
	"hostettler.dev/dnc/models"
)

var (
	TopBarHeight = 6
	TopBarWidth  = LeftColWidth + MidColWidth + RightColWidth + 4

	TopSeparatorWidth = 20

	ColHeight    = 25
	LeftColWidth = 30
	MidColWidth  = 28

	RightColWidth     = 38
	RightContentWidth = RightColWidth - 6

	LongColWidth   = 20
	ColWidth       = 16
	MediumColWidth = 12
	ShortColWidth  = 8
	TinyColWidth   = 3
)

type ScoreScreen struct {
	keymap        KeyMap
	character     *models.Character
	characterInfo *List
	abilities     *List
	skills        *List
	savingThrows  *List
	combatInfo    *List
	attacks       *List
}

func NewScoreScreen(keymap KeyMap, c *models.Character) *ScoreScreen {
	return &ScoreScreen{
		keymap:    keymap,
		character: c,
		characterInfo: NewListWithDefaults().
			WithRows(GetCharacterInfoRows(c)),
		abilities: NewListWithDefaults().
			WithRows(GetAbilityRows(c)),
		skills: NewListWithDefaults().
			WithTitle("Skills").
			WithRows(GetSkillRows(c)).
			SetFocus(true),
		savingThrows: NewListWithDefaults().
			WithTitle("Saving Throws").
			WithRows(GetSavingThrowRows(c)),
		combatInfo: NewListWithDefaults().
			WithTitle("Combat").
			WithRows(GetCombatInfoRows(c)),
		attacks: NewListWithDefaults().
			WithTitle("Attacks").
			WithRows(GetAttackRows(c)),
	}
}

func (s *ScoreScreen) Init() tea.Cmd {
	return nil
}

func (s *ScoreScreen) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return nil, nil
}

func (s *ScoreScreen) View() string {
	characterInfo := s.characterInfo.View()

	abilities := s.abilities.View()

	topBarSeparator := MakeVerticalSeparator(TopBarHeight)

	topBar := DefaultBorderStyle.
		Height(TopBarHeight).
		Width(TopBarWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Center,
			characterInfo,
			lipgloss.PlaceHorizontal(20, lipgloss.Center, topBarSeparator),
			abilities))

	leftColumn := DefaultBorderStyle.
		Height(ColHeight).
		Width(LeftColWidth).
		Render(s.skills.View())

	savingThrows := s.savingThrows.View()

	combatInfo := s.combatInfo.View()

	midBoxInnerSeparator := "\n" +
		GrayTextStyle.Render(strings.Repeat("─", MidColWidth-4)) +
		"\n"

	midColumn := DefaultBorderStyle.
		Width(MidColWidth).
		Height(ColHeight).
		Render(lipgloss.JoinVertical(lipgloss.Center, combatInfo, midBoxInnerSeparator, savingThrows))

	actions := RenderActions(s.character)

	attacks := s.attacks.View()

	rightBoxInnerSeparator := "\n" +
		GrayTextStyle.Render(strings.Repeat("─", RightContentWidth)) +
		"\n"

	rightColumn := DefaultBorderStyle.
		Width(RightColWidth).
		Height(ColHeight).
		Render(lipgloss.JoinVertical(lipgloss.Center, actions, rightBoxInnerSeparator, attacks))

	body := lipgloss.JoinHorizontal(lipgloss.Left, leftColumn, midColumn, rightColumn)

	topSeparator := GrayTextStyle.Render(strings.Repeat("─", lipgloss.Width(body)))

	return lipgloss.JoinVertical(lipgloss.Center, topBar, topSeparator, body)
}

func (s *ScoreScreen) GetCharacterInfoRows() []Row {
	rowCfg := LabeledStringRowConfig{false, ColWidth, 0}
	rows := []Row{
		NewLabeledStringRow(s.keymap, "Name:", &s.character.Name,
			NewStringEditor("Name", &s.character.Name, s.keymap)).WithConfig(rowCfg),
		NewLabeledStringRow(s.keymap, "Levels:", &s.character.ClassLevels,
			NewStringEditor("Levels", &s.character.ClassLevels, s.keymap)).WithConfig(rowCfg),
		NewLabeledStringRow(s.keymap, "Race:", &s.character.Race,
			NewStringEditor("Race", &s.character.Race, s.keymap)).WithConfig(rowCfg),
		NewLabeledStringRow(s.keymap, "Alignment:", &s.character.Alignment,
			NewStringEditor("Alignment", &s.character.Alignment, s.keymap)).WithConfig(rowCfg),
	}
	return rows
}

func (s *ScoreScreen) GetAbilityRows() []Row {
	scorePrinter := func(score int) string {
		return fmt.Sprintf("%3s  ( %+d )", strconv.Itoa(score), models.ToModifier(score))
	}
	rowCfg := LabeledIntRowConfig{scorePrinter, true, ColWidth, ShortColWidth}
	rows := []Row{
		NewLabeledIntRow(s.keymap, "Strength:", &s.character.Abilities.Strength,
			NewIntEditor("Strength", &s.character.Abilities.Strength, s.keymap)).WithConfig(rowCfg),
		NewLabeledIntRow(s.keymap, "Constitution:", &s.character.Abilities.Constitution,
			NewIntEditor("Constitution", &s.character.Abilities.Constitution, s.keymap)).WithConfig(rowCfg),
		NewLabeledIntRow(s.keymap, "Dexterity:", &s.character.Abilities.Dexterity,
			NewIntEditor("Dexterity", &s.character.Abilities.Dexterity, s.keymap)).WithConfig(rowCfg),
		NewLabeledIntRow(s.keymap, "Intelligence:", &s.character.Abilities.Intelligence,
			NewIntEditor("Intelligence", &s.character.Abilities.Intelligence, s.keymap)).WithConfig(rowCfg),
		NewLabeledIntRow(s.keymap, "Wisdom:", &s.character.Abilities.Wisdom,
			NewIntEditor("Wisdom", &s.character.Abilities.Wisdom, s.keymap)).WithConfig(rowCfg),
		NewLabeledIntRow(s.keymap, "Charisma:", &s.character.Abilities.Charisma,
			NewIntEditor("Charisma", &s.character.Abilities.Charisma, s.keymap)).WithConfig(rowCfg),
	}
	return rows
}

func RenderAbility(name string, score int) string {
	scoreStr := fmt.Sprintf("%3s  ( %+d )", strconv.Itoa(score), models.ToModifier(score))
	return RenderEdgeBound(ColWidth, ShortColWidth, name+":", scoreStr)
}

func GetCombatInfoRows(c *models.Character) []Row {
	initiative := models.ToModifier(c.Abilities.Dexterity)
	rows := []Row{
		{RenderEdgeBound(ColWidth, TinyColWidth, "AC", strconv.Itoa(c.ArmorClass))},
		{RenderEdgeBound(ColWidth, TinyColWidth, "Initiative", fmt.Sprintf("%+d", initiative))},
		{RenderEdgeBound(ColWidth, TinyColWidth, "Speed", strconv.Itoa(c.Speed))},
		{RenderEdgeBound(ColWidth-4, 7, "HP", strconv.Itoa(c.CurrentHitPoints)+"/"+strconv.Itoa(c.MaxHitPoints))},
		{RenderEdgeBound(ColWidth-8, TinyColWidth+8, "Hit Dice", c.UsedHitDice+"/"+c.HitDice)},
		{RenderEdgeBound(ColWidth, TinyColWidth, "DS Successes", DeathSaveSymbols(c.DeathSaves.Successes))},
		{RenderEdgeBound(ColWidth, TinyColWidth, "DS Failures", DeathSaveSymbols(c.DeathSaves.Failures))},
	}
	return rows
}

func RenderActions(c *models.Character) string {
	actionTitle := DefaultTextStyle.Render("Actions\n")

	actionBody := DefaultTextStyle.Width(RightContentWidth).Render(c.Actions)

	separator := "\n" +
		GrayTextStyle.Render(strings.Repeat("─", RightContentWidth)) +
		"\n"

	bonusActionTitle := DefaultTextStyle.Render("Bonus Actions\n")

	bonusActionBody := DefaultTextStyle.Width(RightContentWidth).Render(c.BonusActions)

	return lipgloss.JoinVertical(lipgloss.Center, actionTitle, actionBody, separator, bonusActionTitle, bonusActionBody)
}

func GetAttackRows(c *models.Character) []Row {
	rows := []Row{
		(Map(c.Attacks, RenderAttack)),
	}
	for _, a := range c.Attacks {
		rows = append(rows, Row{RenderAttack(a)})
	}
	return rows
}
func GetSkillRows(c *models.Character) []Row {
	rows := []Row{}

	skillFields := structs.Fields(c.Skills)
	for _, field := range skillFields {
		skill := field.Value().(models.Skill)
		mod := skill.ToModifier(c.Abilities, c.ProficiencyBonus)
		bullet := ProficiencySymbol(skill.Proficiency)
		row := Row{RenderEdgeBound(LongColWidth, TinyColWidth, bullet+" "+skill.Name, fmt.Sprintf("%+d", mod))}
		rows = append(rows, row)
	}

	return rows
}

func GetSavingThrowRows(c *models.Character) []Row {
	rows := []Row{}

	skillFields := structs.Fields(c.SavingThrows)
	for _, field := range skillFields {
		saving := field.Value().(models.SavingThrow)
		mod := saving.ToModifier(c.Abilities, c.ProficiencyBonus)
		bullet := ProficiencySymbol(saving.Proficiency)
		row := Row{RenderEdgeBound(ColWidth, TinyColWidth, bullet+" "+saving.Ability, fmt.Sprintf("%+d", mod))}
		rows = append(rows, row)
	}

	return rows
}

func RenderAttack(a models.Attack) string {
	return fmt.Sprintf("%10s %+3d %6s (%s)", a.Name, a.Bonus, a.Damage, a.DamageType)
}

func DeathSaveSymbols(amount int) string {
	return strings.Repeat("●", amount) + strings.Repeat("○", 3-amount)
}

func ProficiencySymbol(p models.ProficiencyLevel) string {
	var bullet string
	switch p {
	case models.NoProficiency:
		bullet = "○"
	case models.Proficient:
		bullet = "◐"
	case models.Expertise:
		bullet = "●"
	}
	return bullet
}
