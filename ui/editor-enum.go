package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type EnumMapping struct {
	Value int
	Label string
}

type EnumEditor struct {
	keymap      KeyMap
	options     []EnumMapping
	value       *int
	cursor      int
	initialized bool
}

func NewEnumEditor(keymap KeyMap, options []EnumMapping, label string, delegatorPointer interface{}) *EnumEditor {
	e := EnumEditor{
		options: options,
	}
	e.Init(keymap, label, delegatorPointer)
	return &e
}

func (e *EnumEditor) Init(keymap KeyMap, label string, delegatorPointer interface{}) {
	e.keymap = keymap
	ptr, ok := delegatorPointer.(*int)
	if !ok {
		panic("Value passed is not a pointer to int")
	}

	e.value = ptr

	for i, opt := range e.options {
		if ptr != nil && opt.Value == *ptr {
			e.cursor = i
			break
		}
	}

	e.initialized = true
}

func (e *EnumEditor) Update(msg tea.Msg) tea.Cmd {
	if !e.initialized {
		return nil
	}

	switch m := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(m, e.keymap.Left, e.keymap.Up):
			e.cursor = (e.cursor - 1 + len(e.options)) % len(e.options)
		case key.Matches(m, e.keymap.Right, e.keymap.Down):
			e.cursor = (e.cursor + 1) % len(e.options)
		}
	}

	return nil
}

func (e *EnumEditor) View() string {
	if !e.initialized || len(e.options) == 0 {
		return ""
	}
	current := e.options[e.cursor]
	return fmt.Sprintf("[ %s ]", current.Label)
}

func (e *EnumEditor) Save() tea.Cmd {
	if e.value != nil && e.cursor < len(e.options) {
		*e.value = e.options[e.cursor].Value
	}
	return nil
}
