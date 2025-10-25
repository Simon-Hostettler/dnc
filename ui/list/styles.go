package list

import (
	"github.com/charmbracelet/lipgloss"
	styles "hostettler.dev/dnc/ui/util"
)

var LeftAlignedListStyle = ListStyles{
	Row:      styles.ItemStyleDefault.Align(lipgloss.Left),
	Selected: styles.ItemStyleSelected.Align(lipgloss.Left),
}
