package list

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"
	"hostettler.dev/dnc/ui/editor"
)

type SeparatorRow struct {
	id     uuid.UUID
	symbol string
	width  int
}

func NewSeparatorRow(symbol string, width int) *SeparatorRow {
	return &SeparatorRow{uuid.New(), symbol, width}
}

func (r *SeparatorRow) Id() uuid.UUID {
	return r.id
}

func (r *SeparatorRow) Init() tea.Cmd {
	return nil
}

func (r *SeparatorRow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return r, nil
}

func (r *SeparatorRow) View() tea.View {
	return tea.NewView(strings.Repeat(r.symbol, r.width))
}

func (c *SeparatorRow) Editors() []editor.ValueEditor {
	return []editor.ValueEditor{}
}

func (c *SeparatorRow) Selectable() bool {
	return false
}
