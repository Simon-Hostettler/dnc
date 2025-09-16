package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type StringEditor struct {
	keymap      KeyMap
	value       *string
	textInput   textinput.Model
	initialized bool
}

func (s *StringEditor) Init(label string, delegatorPointer interface{}, keymap KeyMap) {
	ptr, ok := delegatorPointer.(*string)
	if !ok {
		panic("Value passed is not a pointer to string")
	}
	s.value = ptr

	ti := textinput.New()
	ti.Width = 40

	if ptr != nil {
		ti.SetValue(*ptr)
	}

	s.textInput = ti
	s.initialized = true
}

func (s *StringEditor) Update(msg tea.Msg) tea.Cmd {
	if !s.initialized {
		return nil
	}

	var cmd tea.Cmd
	s.textInput, cmd = s.textInput.Update(msg)
	return cmd
}

func (s *StringEditor) View() string {
	if !s.initialized {
		return ""
	}
	return s.textInput.View()
}

func (s *StringEditor) Save() tea.Cmd {
	if s.value != nil {
		*s.value = s.textInput.Value()
	}
	return nil
}
