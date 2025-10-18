package component

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/util"
)

type SimpleStringComponent struct {
	keymap           util.KeyMap
	name             string
	value            string
	editor           editor.ValueEditor
	focus            bool
	renderName       bool
	highlightOnFocus bool
}

func NewSimpleStringComponent(k util.KeyMap, name string, value string, saveCallback func(string) error, renderName bool, highlightOnFocus bool) *SimpleStringComponent {
	return &SimpleStringComponent{k, name, value, editor.NewStringEditor(k, name, value, saveCallback), false, renderName, highlightOnFocus}
}

func (s *SimpleStringComponent) Init() tea.Cmd {
	return nil
}

func (s *SimpleStringComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keymap.Edit):
			return s, editor.EditValueCmd([]editor.ValueEditor{s.editor})
		}
	}
	return s, nil
}

func (s *SimpleStringComponent) View() string {
	prefix := ""
	if s.renderName {
		prefix = s.name + ": "
	}
	return util.RenderItem(s.focus && s.highlightOnFocus, prefix+s.value)
}

func (s *SimpleStringComponent) Focus() {
	s.focus = true
}

func (s *SimpleStringComponent) Blur() {
	s.focus = false
}

func (s *SimpleStringComponent) InFocus() bool {
	return s.focus
}
