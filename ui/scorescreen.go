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
	cursor    int
	character *models.Character
}

func NewScoreScreen(c *models.Character) *ScoreScreen {
	return &ScoreScreen{
		character: c,
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
	skills := RenderSkills(s.character.Skills, s.character.Abilities, s.character.ProficiencyBonus)
	return lipgloss.JoinVertical(lipgloss.Left, abilities, separator, skills)
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
	modStr := fmt.Sprintf("%+d", models.ToModifier(score))

	innerBorder := HorizontalBorderStyle.Padding(0, 2)
	outerBorder := VerticalBorderStyle.Padding(1, 0).Width(14)

	scoreStr := fmt.Sprintf("%d", score)
	modView := innerBorder.Render(modStr)

	content := lipgloss.JoinVertical(lipgloss.Center, name, "\n"+scoreStr, modView)
	top := outerBorder.Render(content)

	return top

}

func RenderSkills(s models.Skills, a models.Abilities, profBonus int) string {
	rows := []string{}

	skillFields := structs.Fields(s)
	for _, field := range skillFields {
		skill := field.Value().(models.Skill)
		mod := skill.ToModifier(a, profBonus)
		row := fmt.Sprintf("%-18s %3s", skill.Name, fmt.Sprintf("%+d", mod))
		rows = append(rows, row)
	}

	t := lipgloss.JoinVertical(lipgloss.Center, rows...)
	t = VerticalBorderStyle.Padding(1, 0).Width(27).Render(t)
	return t
}
