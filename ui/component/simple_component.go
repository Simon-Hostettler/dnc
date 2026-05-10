package component

import (
	"strconv"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
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
}

func NewSimpleIntComponent(k util.KeyMap, name string, content *int, renderName bool, highlightOnFocus bool) *SimpleComponent[int] {
	return &SimpleComponent[int]{k, name, content, editor.NewIntEditor(k, name, content), false, renderName, highlightOnFocus, strconv.Itoa}
}

func NewSimpleStringComponent(k util.KeyMap, name string, content *string, renderName bool, highlightOnFocus bool) *SimpleComponent[string] {
	return &SimpleComponent[string]{k, name, content, editor.NewStringEditor(k, name, content), false, renderName, highlightOnFocus, func(s string) string { return s }}
}

func (s *SimpleComponent[T]) Init() tea.Cmd {
	return nil
}

func (s *SimpleComponent[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
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
