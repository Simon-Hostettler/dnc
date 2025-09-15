package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// implements ValueEditor[string]
type StringEditor struct {
	delegate    *string
	textInput   textinput.Model
	initialized bool
}

func (s *StringEditor) Init(label string, delegate *string) {
	s.delegate = delegate

	ti := textinput.New()
	ti.Placeholder = label
	ti.Width = 40

	if delegate != nil {
		ti.SetValue(*delegate)
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
	if s.delegate != nil {
		*s.delegate = s.textInput.Value()
	}
	return nil
}
