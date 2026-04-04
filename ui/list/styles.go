package list

import (
	"charm.land/lipgloss/v2"
	"hostettler.dev/dnc/ui/styles"
)

var LeftAlignedListStyle = ListStyles{
	Row:      styles.ItemStyleDefault.Align(lipgloss.Left),
	Selected: styles.ItemStyleSelected.Align(lipgloss.Left),
}
