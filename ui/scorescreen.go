package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/structs"
	"hostettler.dev/dnc/models"
)

type ScoreScreen struct {
	character    *models.Character
	skills       *Table
	savingThrows *Table
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
	separator := GrayTextStyle.Render(strings.Repeat("─", lipgloss.Width(abilities)))
	skills := VerticalBorderStyle.Render(s.skills.View())
	savingThrows := VerticalBorderStyle.Render(s.savingThrows.View())
	body := lipgloss.JoinHorizontal(lipgloss.Left, skills, savingThrows)
	return lipgloss.JoinVertical(lipgloss.Left, abilities, separator, body)
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

	innerBorder := HorizontalBorderStyle.Padding(0, 2)
	outerBorder := VerticalBorderStyle

	scoreStr := DefaultTextStyle.Render(fmt.Sprintf("%d", score))
	modView := innerBorder.Render(modStr)

	content := lipgloss.JoinVertical(lipgloss.Center, DefaultTextStyle.Render(name), "\n"+scoreStr, modView)
	top := outerBorder.Render(content)

	return top

}

func SkillsToRows(c *models.Character) []Row {
	rows := []Row{}

	skillFields := structs.Fields(c.Skills)
	for _, field := range skillFields {
		skill := field.Value().(models.Skill)
		mod := skill.ToModifier(c.Abilities, c.ProficiencyBonus)
		row := Row{fmt.Sprintf("%-18s %3s", skill.Name, fmt.Sprintf("%+d", mod))}
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
		row := Row{fmt.Sprintf("%-14s %3s", saving.Ability, fmt.Sprintf("%+d", mod))}
		rows = append(rows, row)
	}

	return rows
}
