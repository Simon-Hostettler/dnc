package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type StringEditor struct {
	keymap      KeyMap
	label       string
	value       *string
	textInput   textinput.Model
	initialized bool
	focus       bool
}

func NewStringEditor(keymap KeyMap, label string, delegatorPointer interface{}) *StringEditor {
	s := StringEditor{}
	s.Init(keymap, label, delegatorPointer)
	return &s
}

func (s *StringEditor) Init(keymap KeyMap, label string, delegatorPointer interface{}) {
	ptr, ok := delegatorPointer.(*string)
	if !ok {
		panic("Value passed is not a pointer to string")
	}
	s.value = ptr

	ti := textinput.New()
	ti.Prompt = ""

	if ptr != nil {
		ti.SetValue(*ptr)
	}

	s.textInput = ti
	s.label = label
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
	return RenderItem(s.focus, s.label+":") + " " + ItemStyleDefault.Render(s.textInput.View())
}

func (s *StringEditor) Save() tea.Cmd {
	if s.value != nil {
		*s.value = s.textInput.Value()
	}
	return nil
}

func (e *StringEditor) Focus() {
	e.textInput.Focus()
	e.focus = true
}

func (e *StringEditor) Blur() {
	e.textInput.Blur()
	e.focus = false
}
