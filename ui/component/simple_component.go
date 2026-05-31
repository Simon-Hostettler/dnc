package component

import (
	"strconv"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

type SimpleComponent[T any] struct {
	keymap           util.KeyMap
	name             string
	content          *T
	editor           editor.ValueEditor
	focus            bool
	renderName       bool
	highlightOnFocus bool
	format           func(T) string
	cycleAction      func(*T) tea.Cmd
}

func NewSimpleIntComponent(k util.KeyMap, name string, content *int, renderName bool, highlightOnFocus bool) *SimpleComponent[int] {
	return &SimpleComponent[int]{k, name, content, editor.NewIntEditor(k, name, content), false, renderName, highlightOnFocus, strconv.Itoa, nil}
}

func NewSimpleStringComponent(k util.KeyMap, name string, content *string, renderName bool, highlightOnFocus bool) *SimpleComponent[string] {
	return &SimpleComponent[string]{k, name, content, editor.NewStringEditor(k, name, content), false, renderName, highlightOnFocus, func(s string) string { return s }, nil}
}

// Renders an int-backed enum using the given symbols and
// cycles to the next value when the Cycle key is pressed.
func NewSimpleEnumComponent(k util.KeyMap, name string, content *int, symbols []styles.EnumMapping, renderName bool, highlightOnFocus bool) *SimpleComponent[int] {
	format := func(v int) string {
		for _, m := range symbols {
			if m.Value == v {
				return m.Label
			}
		}
		return ""
	}
	cycle := func(v *int) tea.Cmd {
		*v = (*v + 1) % len(symbols)
		return command.WriteBackRequest
	}
	return &SimpleComponent[int]{k, name, content, editor.NewEnumEditor(k, symbols, name, content), false, renderName, highlightOnFocus, format, cycle}
}

func (s *SimpleComponent[T]) WithFormat(format func(T) string) *SimpleComponent[T] {
	s.format = format
	return s
}

func (s *SimpleComponent[T]) WithCycleAction(action func(*T) tea.Cmd) *SimpleComponent[T] {
	s.cycleAction = action
	return s
}

func (s *SimpleComponent[T]) Init() tea.Cmd {
	return nil
}

func (s *SimpleComponent[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, s.keymap.Cycle) && s.cycleAction != nil:
			return s, s.cycleAction(s.content)
		case key.Matches(msg, s.keymap.Edit):
			return s, editor.EditValueCmd([]editor.ValueEditor{s.editor})
		}
	}
	return s, nil
}

func (s *SimpleComponent[T]) View() tea.View {
	prefix := ""
	if s.renderName {
		prefix = s.name + ": "
	}
	return tea.NewView(styles.RenderItem(s.focus && s.highlightOnFocus, prefix+s.format(*s.content)))
}

func (s *SimpleComponent[T]) Focus() {
	s.focus = true
}

func (s *SimpleComponent[T]) Blur() {
	s.focus = false
}

func (s *SimpleComponent[T]) InFocus() bool {
	return s.focus
}
