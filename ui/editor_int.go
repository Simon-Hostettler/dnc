package ui

import (
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type IntEditor struct {
	keymap      KeyMap
	label       string
	value       *int
	textInput   textinput.Model
	initialized bool
	focus       bool
}

func NewIntEditor(keymap KeyMap, label string, delegatorPointer interface{}) *IntEditor {
	s := IntEditor{}
	s.Init(keymap, label, delegatorPointer)
	return &s
}

func (s *IntEditor) Init(keymap KeyMap, label string, delegatorPointer interface{}) {
	ptr, ok := delegatorPointer.(*int)
	if !ok {
		panic("Value passed is not a pointer to int")
	}
	s.value = ptr

	ti := textinput.New()
	ti.Prompt = ""

	if ptr != nil {
		ti.SetValue(strconv.Itoa(*ptr))
	}

	s.textInput = ti
	s.label = label
	s.initialized = true
}

func (s *IntEditor) Update(msg tea.Msg) tea.Cmd {
	if !s.initialized {
		return nil
	}

	var cmd tea.Cmd
	s.textInput, cmd = s.textInput.Update(msg)
	return cmd
}

func (s *IntEditor) View() string {
	if !s.initialized {
		return ""
	}
	return RenderItem(s.focus, s.label+":") + " " + ItemStyleDefault.Render(s.textInput.View())
}

func (s *IntEditor) Save() tea.Cmd {
	value, err := strconv.Atoi(s.textInput.Value())
	if err != nil {
		return nil
	}
	if s.value != nil {
		*s.value = value
	}
	return nil
}

func (e *IntEditor) Focus() {
	e.textInput.Focus()
	e.focus = true
}

func (e *IntEditor) Blur() {
	e.textInput.Blur()
	e.focus = false
}
