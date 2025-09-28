package list

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/ui/editor"
)

type SeparatorRow struct {
	symbol string
	width  int
}

func NewSeparatorRow(symbol string, width int) *SeparatorRow {
	return &SeparatorRow{symbol, width}
}

func (r *SeparatorRow) Init() tea.Cmd {
	return nil
}

func (r *SeparatorRow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return r, nil
}

func (r *SeparatorRow) View() string {
	return strings.Repeat(r.symbol, r.width)
}

func (c *SeparatorRow) Editors() []editor.ValueEditor {
	return []editor.ValueEditor{}
}
