package list

import (
	"github.com/charmbracelet/lipgloss"

	"hostettler.dev/dnc/ui/util"
)

var (
	LeftAlignedListStyle = ListStyles{
		Row:      util.ItemStyleDefault.Align(lipgloss.Left),
		Selected: util.ItemStyleSelected.Align(lipgloss.Left),
	}
)
