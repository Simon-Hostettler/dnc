package list

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/ui/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/util"
)

/*
The appender will simply send out an AppendElementCmd,
the implementation is the client's responsibility.
*/
type AppenderRow struct {
	keymap util.KeyMap
	tag    string
}

func NewAppenderRow(keymap util.KeyMap, tag string) *AppenderRow {
	return &AppenderRow{keymap, tag}
}

func (r *AppenderRow) Init() tea.Cmd {
	return nil
}

func (r *AppenderRow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, r.keymap.Select):
			return r, command.AppendElementCmd(r.tag)
		}
	}
	return r, nil
}

func (r *AppenderRow) View() string {
	return "[ + ]"
}

func (c *AppenderRow) Editors() []editor.ValueEditor {
	return []editor.ValueEditor{}
}
