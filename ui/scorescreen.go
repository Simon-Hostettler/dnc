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
	StandardBoxWidth  = 30
	StandardBoxHeight = 25
	LongColWidth      = 20
	ColWidth          = 16
	ShortColWidth     = 3
)

type ScoreScreen struct {
	character    *models.Character
	skills       *Table
	savingThrows *Table
	combatInfo   *Table
	attacks      *Table
}

func NewScoreScreen(c *models.Character) *ScoreScreen {
	return &ScoreScreen{
		character: c,
		skills: NewTableWithDefaults().
			WithTitle("Skills").
			WithRows(SkillsToRows(c)).
			SetFocus(true),
		savingThrows: NewTableWithDefaults().
			WithTitle("Saving Throws").
			WithRows(SavingThrowsToRows(c)),
		combatInfo: NewTableWithDefaults().
			WithTitle("Combat").
			WithRows(GetCombatInfoRows(c)),
		attacks: NewTableWithDefaults().
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
	abilities := RenderAbilities(s.character.Abilities)

	savingThrows := s.savingThrows.View()

	combatInfo := s.combatInfo.View()

	midBoxInnerSeparator := "\n" +
		GrayTextStyle.Render(strings.Repeat("─", StandardBoxWidth-6)) +
		"\n"

	midColumn := DefaultBorderStyle.
		Width(StandardBoxWidth - 2).
		Height(StandardBoxHeight).
		Render(lipgloss.JoinVertical(lipgloss.Center, combatInfo, midBoxInnerSeparator, savingThrows))

	actions := RenderActions(s.character)

	attacks := s.attacks.View()

	rightBoxInnerSeparator := "\n" +
		GrayTextStyle.Render(strings.Repeat("─", StandardBoxWidth+2)) +
		"\n"

	rightColumn := DefaultBorderStyle.
		Width(StandardBoxWidth + 8).
		Height(StandardBoxHeight).
		Render(lipgloss.JoinVertical(lipgloss.Center, actions, rightBoxInnerSeparator, attacks))

	skills := DefaultBorderStyle.
		Height(lipgloss.Height(midColumn) - 2).
		Width(StandardBoxWidth).
		Height(StandardBoxHeight).
		Render(s.skills.View())

	body := lipgloss.JoinHorizontal(lipgloss.Left, skills, midColumn, rightColumn)

	topSeparator := GrayTextStyle.Render(strings.Repeat("─", lipgloss.Width(body)))

	return lipgloss.JoinVertical(lipgloss.Center, abilities, topSeparator, body)
}

func RenderAbilities(a models.Abilities) string {
	strength := RenderAbility("Strength", a.Strength)
	constitution := RenderAbility("Constitution", a.Constitution)
	dexterity := RenderAbility("Dexterity", a.Dexterity)
	intelligence := RenderAbility("Intelligence", a.Intelligence)
	wisdom := RenderAbility("Wisdom", a.Wisdom)
	charisma := RenderAbility("Charisma", a.Charisma)

	return lipgloss.JoinHorizontal(lipgloss.Center, strength, constitution, dexterity, intelligence, wisdom, charisma)
}

func RenderAbility(name string, score int) string {
	modStr := DefaultTextStyle.Render(fmt.Sprintf("%+d", models.ToModifier(score)))

	innerBorder := DefaultBorderStyle.Padding(0, 2)
	outerBorder := DefaultBorderStyle.Padding(1, 0, 0).Width(14)

	scoreStr := DefaultTextStyle.Render(fmt.Sprintf("%d", score))
	modView := innerBorder.Render(modStr)

	content := lipgloss.JoinVertical(lipgloss.Center, DefaultTextStyle.Render(name), "\n"+scoreStr, modView)
	top := outerBorder.Render(content)

	return top

}

func GetCombatInfoRows(c *models.Character) []Row {
	initiative := models.ToModifier(c.Abilities.Dexterity)
	rows := []Row{
		{
			RenderEdgeBound(ColWidth, ShortColWidth, "AC", strconv.Itoa(c.ArmorClass)),
		},
		{
			RenderEdgeBound(ColWidth, ShortColWidth, "Initiative", fmt.Sprintf("%+d", initiative)),
		},
		{
			RenderEdgeBound(ColWidth, ShortColWidth, "Speed", strconv.Itoa(c.Speed)),
		},
		{
			RenderEdgeBound(ColWidth-4, 7, "HP", strconv.Itoa(c.CurrentHitPoints)+"/"+strconv.Itoa(c.MaxHitPoints)),
		},
		{
			RenderEdgeBound(ColWidth-8, ShortColWidth+8, "Hit Dice", c.UsedHitDice+"/"+c.HitDice),
		},
		{
			RenderEdgeBound(ColWidth, ShortColWidth, "DS Successes", DeathSaveSymbols(c.DeathSaves.Successes)),
		},
		{
			RenderEdgeBound(ColWidth, ShortColWidth, "DS Failures", DeathSaveSymbols(c.DeathSaves.Failures)),
		},
	}
	return rows
}

func RenderActions(c *models.Character) string {
	actionTitle := DefaultTextStyle.Render("Actions\n")

	actionBody := DefaultTextStyle.Width(StandardBoxWidth + 2).Render(c.Actions)

	separator := "\n" +
		GrayTextStyle.Render(strings.Repeat("─", StandardBoxWidth+2)) +
		"\n"

	bonusActionTitle := DefaultTextStyle.Render("Bonus Actions\n")

	bonusActionBody := DefaultTextStyle.Width(StandardBoxWidth + 2).Render(c.BonusActions)

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
func SkillsToRows(c *models.Character) []Row {
	rows := []Row{}

	skillFields := structs.Fields(c.Skills)
	for _, field := range skillFields {
		skill := field.Value().(models.Skill)
		mod := skill.ToModifier(c.Abilities, c.ProficiencyBonus)
		bullet := ProficiencySymbol(skill.Proficiency)
		row := Row{RenderEdgeBound(LongColWidth, ShortColWidth, bullet+" "+skill.Name, fmt.Sprintf("%+d", mod))}
		rows = append(rows, row)
	}

	return rows
}

func SavingThrowsToRows(c *models.Character) []Row {
	rows := []Row{}

	skillFields := structs.Fields(c.SavingThrows)
	for _, field := range skillFields {
		saving := field.Value().(models.SavingThrow)
		mod := saving.ToModifier(c.Abilities, c.ProficiencyBonus)
		bullet := ProficiencySymbol(saving.Proficiency)
		row := Row{RenderEdgeBound(ColWidth, ShortColWidth, bullet+" "+saving.Ability, fmt.Sprintf("%+d", mod))}
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
