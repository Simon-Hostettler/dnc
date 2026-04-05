package component

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

type SimpleStringComponent struct {
	keymap           util.KeyMap
	name             string
	content          *string
	editor           editor.ValueEditor
	focus            bool
	renderName       bool
	highlightOnFocus bool
}

func NewSimpleStringComponent(k util.KeyMap, name string, content *string, renderName bool, highlightOnFocus bool) *SimpleStringComponent {
	return &SimpleStringComponent{k, name, content, editor.NewStringEditor(k, name, content), false, renderName, highlightOnFocus}
}

func (s *SimpleStringComponent) Init() tea.Cmd {
	return nil
}

func (s *SimpleStringComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, s.keymap.Edit):
			return s, editor.EditValueCmd([]editor.ValueEditor{s.editor})
		}
	}
	return s, nil
}

func (s *SimpleStringComponent) View() tea.View {
	prefix := ""
	if s.renderName {
		prefix = s.name + ": "
	}
	return tea.NewView(styles.RenderItem(s.focus && s.highlightOnFocus, prefix+*s.content))
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
