package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	DefaultTextStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	AppTitleStyle     = lipgloss.NewStyle().Background(lipgloss.Color("212")).Padding(0, 1)
	ItemStyleSelected = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	ItemStyleDefault  = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	CenteredStyle     = lipgloss.NewStyle().Align(lipgloss.Center)
	MainBorderStyle   = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("7"))
	RoundedBorderNoBottomStyle = lipgloss.NewStyle().Border(RoundedBorder, true, true, false).
					BorderForeground(lipgloss.Color("7")).Align(lipgloss.Center)

	RoundedBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "└",
		BottomRight: "┘",
	}
)

func RenderList(l []string, selected int) string {
	s := ""

	for i, el := range l {
		if i == selected {
			s += "> " + ItemStyleSelected.Render(el) + "\n"
		} else {
			s += "• " + ItemStyleDefault.Render(el) + "\n"
		}
	}
	return s
}
