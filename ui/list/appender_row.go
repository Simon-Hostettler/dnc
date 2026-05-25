package list

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/util"
)

// Renders the "[ + ]" affordance at the end of a section.
// On Select, it invokes onAppend;
type AppenderRow struct {
	keymap   util.KeyMap
	onAppend func() tea.Cmd
}

func NewAppenderRow(keymap util.KeyMap, onAppend func() tea.Cmd) *AppenderRow {
	return &AppenderRow{keymap, onAppend}
}

func (r *AppenderRow) Init() tea.Cmd {
	return nil
}

func (r *AppenderRow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok && key.Matches(k, r.keymap.Select) {
		return r, r.Trigger()
	}
	return r, nil
}

func (r *AppenderRow) Trigger() tea.Cmd {
	if r.onAppend == nil {
		return nil
	}
	return r.onAppend()
}

func (r *AppenderRow) View() tea.View {
	return tea.NewView("[ + ]")
}

func (c *AppenderRow) Editors() []editor.ValueEditor {
	return []editor.ValueEditor{}
}

func (r *AppenderRow) Selectable() bool {
	return true
}
