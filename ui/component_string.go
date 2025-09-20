package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type SimpleStringComponent struct {
	keymap  KeyMap
	name    string
	content *string
	editor  ValueEditor
	focus   bool
}

func NewSimpleStringComponent(k KeyMap, name string, content *string) *SimpleStringComponent {
	return &SimpleStringComponent{k, name, content, NewStringEditor(k, name, content), false}
}

func (s *SimpleStringComponent) Init() tea.Cmd {
	return nil
}

func (s *SimpleStringComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keymap.Edit):
			return s, EditValueCmd([]ValueEditor{s.editor})
		}

	}
	return s, nil
}

func (s *SimpleStringComponent) View() string {
	return *s.content
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
