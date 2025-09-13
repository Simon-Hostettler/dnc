package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
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
	v := RenderAbilities(s.character.Abilities)
	v += "\n"
	v += RenderSkills(s.character.Skills, s.character.Abilities, s.character.ProficiencyBonus)
	return v
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

	innerBorder := RoundedBorderNoBottomStyle.Padding(0, 2)
	outerBorder := RoundedBorderNoBottomStyle.Width(14)

	scoreStr := fmt.Sprintf("%d", score)
	modView := innerBorder.Render(modStr)

	content := lipgloss.JoinVertical(lipgloss.Center, name+"\n", scoreStr, modView)
	top := outerBorder.Render(content)
	bottom := DefaultTextStyle.Render("└───┴──────┴───┘")

	return lipgloss.JoinVertical(lipgloss.Center, top, bottom)

}

func RenderSkills(s models.Skills, a models.Abilities, profBonus int) string {
	columns := []table.Column{
		{Title: "Skill", Width: 18},
		{Title: "Modifier", Width: 10},
	}
	rows := []table.Row{}

	skillFields := structs.Fields(s)
	for _, field := range skillFields {
		skill := field.Value().(models.Skill)
		mod := skill.ToModifier(a, profBonus)
		rows = append(rows, table.Row{
			skill.Name,
			fmt.Sprintf("%+d", mod),
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)
	st := table.DefaultStyles()
	st.Header = st.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	st.Selected = st.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(st)
	return t.View()
}
