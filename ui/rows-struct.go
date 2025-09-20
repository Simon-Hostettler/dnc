package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type StructRow[T any] struct {
	keymap     KeyMap
	value      *T
	renderFunc func(val *T) string
	editors    []ValueEditor
}

func NewStructRow[T any](
	keymap KeyMap,
	value *T,
	renderFunc func(val *T) string,
	editors []ValueEditor,
) *StructRow[T] {
	return &StructRow[T]{
		keymap:     keymap,
		value:      value,
		renderFunc: renderFunc,
		editors:    editors,
	}
}

func (r *StructRow[T]) Init() tea.Cmd {
	return nil
}

func (r *StructRow[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, r.keymap.Edit) {
			return r, EditValueCmd(r.editors)
		}
	}
	return r, nil
}

func (r *StructRow[T]) View() string {
	return r.renderFunc(r.value)
}

func (r *StructRow[T]) Editors() []ValueEditor {
	return r.editors
}
