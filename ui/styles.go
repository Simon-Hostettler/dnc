package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	HighlightColor  = lipgloss.Color("#7D56F4")
	TextColor       = lipgloss.Color("#FAFAFA")
	SecondaryColor  = lipgloss.Color("8")
	BackgroundColor = lipgloss.Color("0")

	DefaultTextStyle = lipgloss.NewStyle().Foreground(TextColor)
	FlippedText      = lipgloss.NewStyle().Background(SecondaryColor).Foreground(BackgroundColor)
	GrayTextStyle    = lipgloss.NewStyle().Foreground(SecondaryColor)

	ItemStyleSelected = lipgloss.NewStyle().Background(HighlightColor)
	ItemStyleDefault  = lipgloss.NewStyle().Foreground(TextColor)

	CenteredStyle = lipgloss.NewStyle().Align(lipgloss.Center)

	MainBorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(TextColor)
	RoundedBorderNoBottomStyle = lipgloss.NewStyle().Border(RoundedBorder, true, true, false).
					BorderForeground(SecondaryColor).Align(lipgloss.Center)
	VerticalBorderStyle = lipgloss.NewStyle().Border(VerticalBorder).
				BorderForeground(SecondaryColor).Align(lipgloss.Center)
	HorizontalBorderStyle = lipgloss.NewStyle().Border(HorizontalBorder).
				BorderForeground(SecondaryColor).Align(lipgloss.Center)
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
