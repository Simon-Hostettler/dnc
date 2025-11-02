package list

import (
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/ui/styles"
)

var LeftAlignedListStyle = ListStyles{
	Row:      styles.ItemStyleDefault.Align(lipgloss.Left),
	Selected: styles.ItemStyleSelected.Align(lipgloss.Left),
}
