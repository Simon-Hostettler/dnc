package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	SmallScreenWidth = 60
	ScreenWidth      = 100

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

	MainBorderStyle = lipgloss.
			NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(TextColor)
	RoundedBorderNoBottomStyle = lipgloss.
					NewStyle().
					Border(RoundedBorder, true, true, false).
					BorderForeground(SecondaryColor).
					Align(lipgloss.Center)
	DefaultBorderStyle = lipgloss.
				NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(SecondaryColor).
				Align(lipgloss.Center).
				Padding(1, 2)
	NoBorderStyle = lipgloss.
			NewStyle().
			Border(RoundedBorder, false, false, false)

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

func ForceWidth(text string, width int) string {
	return lipgloss.NewStyle().Width(width).Render(text)
}

func WithPadding(text string, left int, right int, top int, bottom int) string {
	return lipgloss.NewStyle().Padding(top, right, bottom, left).Render(text)
}

func RenderItem(selected bool, item string) string {
	if selected {
		return ItemStyleSelected.Render(item)
	} else {
		return ItemStyleDefault.Render(item)
	}
}

func MakeVerticalSeparator(height int) string {
	bars := make([]string, height)
	for i := range bars {
		bars[i] = "│"
	}
	return GrayTextStyle.Render(lipgloss.JoinVertical(lipgloss.Center, bars...))
}

func MakeHorizontalSeparator(width int, padding int) string {
	return GrayTextStyle.Padding(padding, 0).Render(strings.Repeat("─", width))
}

func RenderEdgeBound(w1 int, w2 int, str1 string, str2 string) string {
	format := fmt.Sprintf("%%-%ds%%%ds", w1, w2)
	return fmt.Sprintf(format, str1, str2)
}

func RenderLeftBound(w1 int, str1 string, str2 string) string {
	format := fmt.Sprintf("%%-%ds %%s", w1)
	return fmt.Sprintf(format, str1, str2)
}

func ForceLineBreaks(s string, w int) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		b.WriteRune(r)
		if (i+1)%w == 0 && i != len(runes)-1 {
			b.WriteRune('\n')
		}
	}
	return b.String()
}

func PrettySpellSlots(used int, max int) string {
	if max <= 0 {
		return "∅"
	}
	s := strings.Repeat("■", used)
	s += strings.Repeat("□", max-used)
	return DefaultTextStyle.Render(s)
}

func PrettyAttunementSlots(used int) string {
	if used == 0 {
		return ""
	} else {
		s := strings.Repeat("■", used)
		s += strings.Repeat("□", 3-used)
		return s
	}
}

func WithSign(i int) string { return fmt.Sprintf("%+d", i) }

func PrettyDeathSaves(amount int) string {
	return strings.Repeat("●", amount) + strings.Repeat("○", 3-amount)
}

func PrettyBool(b bool) string {
	if b {
		return "■"
	} else {
		return "□"
	}
}

func PrettyBoolCircle(b bool) string {
	if b {
		return "●"
	} else {
		return "○"
	}
}
