package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	AppTitleStyle     = lipgloss.NewStyle().Background(lipgloss.Color("212")).Padding(0, 1)
	ItemStyleSelected = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	ItemStyleDefault  = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	MainBorderStyle   = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("7"))
)
