package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	DefaultTextStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	FlippedText       = lipgloss.NewStyle().Background(lipgloss.Color("8")).Foreground(lipgloss.Color("0"))
	GrayTextStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	AppTitleStyle     = lipgloss.NewStyle().Background(lipgloss.Color("#7D56F4")).Padding(0, 1)
	ItemStyleSelected = lipgloss.NewStyle().Background(lipgloss.Color("#7D56F4"))
	ItemStyleDefault  = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	CenteredStyle     = lipgloss.NewStyle().Align(lipgloss.Center)
	MainBorderStyle   = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("7"))
	RoundedBorderNoBottomStyle = lipgloss.NewStyle().Border(RoundedBorder, true, true, false).
					BorderForeground(lipgloss.Color("8")).Align(lipgloss.Center)
	VerticalBorderStyle = lipgloss.NewStyle().Border(VerticalBorder).
				BorderForeground(lipgloss.Color("8")).Align(lipgloss.Center)
	HorizontalBorderStyle = lipgloss.NewStyle().Border(HorizontalBorder).
				BorderForeground(lipgloss.Color("8")).Align(lipgloss.Center)
	NoBorderStyle = lipgloss.NewStyle().Border(RoundedBorder, false, false, false)

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

	VerticalBorder = lipgloss.Border{
		Top:         " ",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     " ",
		TopRight:    " ",
		BottomLeft:  " ",
		BottomRight: " ",
	}

	HorizontalBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        " ",
		Right:       " ",
		TopLeft:     " ",
		TopRight:    " ",
		BottomLeft:  " ",
		BottomRight: " ",
	}
)

func RenderList(l []string, selected int) string {
	s := ""

	for i, el := range l {
		if i == selected {
			s += ItemStyleSelected.Render(el) + "\n"
		} else {
			s += ItemStyleDefault.Render(el) + "\n"
		}
	}
	return s
}
