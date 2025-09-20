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

func GetCharacterInfoRows(k KeyMap, c *models.Character) []Row {
	rowCfg := LabeledStringRowConfig{false, ColWidth, 0}
	rows := []Row{
		NewLabeledStringRow(k, "Name:", &c.Name,
			NewStringEditor(k, "Name", &c.Name)).WithConfig(rowCfg),
		NewLabeledStringRow(k, "Levels:", &c.ClassLevels,
			NewStringEditor(k, "Levels", &c.ClassLevels)).WithConfig(rowCfg),
		NewLabeledStringRow(k, "Race:", &c.Race,
			NewStringEditor(k, "Race", &c.Race)).WithConfig(rowCfg),
		NewLabeledStringRow(k, "Alignment:", &c.Alignment,
			NewStringEditor(k, "Alignment", &c.Alignment)).WithConfig(rowCfg),
	}
	return rows
}

func GetAbilityRows(k KeyMap, c *models.Character) []Row {
	scorePrinter := func(score int) string {
		return fmt.Sprintf("%3s  ( %+d )", strconv.Itoa(score), models.ToModifier(score))
	}
	rowCfg := LabeledIntRowConfig{scorePrinter, true, ColWidth, ShortColWidth}
	rows := []Row{
		NewLabeledIntRow(k, "Strength:", &c.Abilities.Strength,
			NewIntEditor(k, "Strength", &c.Abilities.Strength)).WithConfig(rowCfg),
		NewLabeledIntRow(k, "Constitution:", &c.Abilities.Constitution,
			NewIntEditor(k, "Constitution", &c.Abilities.Constitution)).WithConfig(rowCfg),
		NewLabeledIntRow(k, "Dexterity:", &c.Abilities.Dexterity,
			NewIntEditor(k, "Dexterity", &c.Abilities.Dexterity)).WithConfig(rowCfg),
		NewLabeledIntRow(k, "Intelligence:", &c.Abilities.Intelligence,
			NewIntEditor(k, "Intelligence", &c.Abilities.Intelligence)).WithConfig(rowCfg),
		NewLabeledIntRow(k, "Wisdom:", &c.Abilities.Wisdom,
			NewIntEditor(k, "Wisdom", &c.Abilities.Wisdom)).WithConfig(rowCfg),
		NewLabeledIntRow(k, "Charisma:", &c.Abilities.Charisma,
			NewIntEditor(k, "Charisma", &c.Abilities.Charisma)).WithConfig(rowCfg),
	}
	return rows
}

func GetCombatInfoRows(k KeyMap, c *models.Character) []Row {
	standardCfg := LabeledIntRowConfig{strconv.Itoa, true, ColWidth, TinyColWidth}
	dsConfig := LabeledIntRowConfig{DeathSaveSymbols, true, ColWidth, TinyColWidth}
	rows := []Row{
		NewLabeledIntRow(k, "AC", &c.ArmorClass,
			NewIntEditor(k, "AC", &c.ArmorClass)).WithConfig(standardCfg),
		NewLabeledIntRow(k, "Initiative", &c.Initiative,
			NewIntEditor(k, "Initiative", &c.Initiative)).
			WithConfig(LabeledIntRowConfig{func(i int) string { return fmt.Sprintf("%+d", i) }, true, ColWidth, TinyColWidth}),
		NewLabeledIntRow(k, "Speed", &c.Speed,
			NewIntEditor(k, "Speed", &c.Speed)).WithConfig(standardCfg),
		NewStructRow(k, &HPInfo{&c.CurrentHitPoints, &c.MaxHitPoints}, renderHPInfoRow,
			[]ValueEditor{NewIntEditor(k, "Current HP", &c.CurrentHitPoints),
				NewIntEditor(k, "Max HP", &c.MaxHitPoints)}),
		NewStructRow(k, &HitDiceInfo{&c.UsedHitDice, &c.HitDice}, renderHitDiceInfoRow,
			[]ValueEditor{NewIntEditor(k, "Used Hit Dice", &c.UsedHitDice),
				NewIntEditor(k, "Hit Dice", &c.HitDice)}),
		NewLabeledIntRow(k, "DS Successes", &c.DeathSaves.Successes,
			NewIntEditor(k, "DS Successes", &c.DeathSaves.Successes)).WithConfig(dsConfig),
		NewLabeledIntRow(k, "DS Failures", &c.DeathSaves.Failures,
			NewIntEditor(k, "DS Failures", &c.DeathSaves.Failures)).WithConfig(dsConfig),
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

// screen specific types + utility functions

type HPInfo struct {
	current *int
	max     *int
}

func renderHPInfoRow(hp *HPInfo) string {
	return RenderEdgeBound(ColWidth-4, 7, "HP", strconv.Itoa(*hp.current)+"/"+strconv.Itoa(*hp.max))
}

type HitDiceInfo struct {
	current *string
	max     *string
}

func renderHitDiceInfoRow(hd *HitDiceInfo) string {
	return RenderEdgeBound(ShortColWidth, MediumColWidth, "Hit Dice", c.UsedHitDice+"/"+c.HitDice)
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
