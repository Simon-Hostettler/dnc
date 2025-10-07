package list

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/ui/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/util"
)

type StructRow[T any] struct {
	keymap     util.KeyMap
	value      *T
	renderFunc func(val *T) string
	editors    []editor.ValueEditor
	originator command.ScreenIndex
	destructor func() tea.Cmd
}

func NewStructRow[T any](
	keymap util.KeyMap,
	value *T,
	renderFunc func(val *T) string,
	editors []editor.ValueEditor,
) *StructRow[T] {
	return &StructRow[T]{
		keymap:     keymap,
		value:      value,
		renderFunc: renderFunc,
		editors:    editors,
	}
}

func (r *StructRow[T]) WithDestructor(caller command.ScreenIndex, callback func() tea.Cmd) *StructRow[T] {
	r.originator = caller
	r.destructor = callback
	return r
}

func (r *StructRow[T]) Init() tea.Cmd {
	return nil
}

func (r *StructRow[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, r.keymap.Edit):
			return r, editor.EditValueCmd(r.editors)
		case key.Matches(msg, r.keymap.Delete) && r.destructor != nil:
			return r, command.LaunchConfirmationDialogueCmd(r.destructor)
		}
	}
	return r, nil
}

func (r *StructRow[T]) View() string {
	return r.renderFunc(r.value)
}

func (r *StructRow[T]) Editors() []editor.ValueEditor {
	return r.editors
}

func (r *StructRow[T]) Value() *T {
	return r.value
}
