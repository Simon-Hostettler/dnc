package ui

import (
	"fmt"
	"reflect"

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
	value       reflect.Value
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

	ptrValue := reflect.ValueOf(delegatorPointer)
	if ptrValue.Kind() != reflect.Ptr || !ptrValue.Elem().IsValid() {
		panic("Value passed is not a valid pointer")
	}

	elem := ptrValue.Elem()
	kind := elem.Kind()
	if kind < reflect.Int || kind > reflect.Int64 {
		panic(fmt.Sprintf("Value passed is not a pointer to int-like, got: %s", kind))
	}

	e.value = ptrValue

	currentValue := int(elem.Int())
	for i, opt := range e.options {
		if opt.Value == currentValue {
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
	if e.value.IsValid() && e.cursor < len(e.options) {
		e.value.Elem().SetInt(int64(e.options[e.cursor].Value))
	}
	return nil
}
