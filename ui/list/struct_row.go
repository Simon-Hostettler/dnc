package list

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/util"
)

type StructRow[T any] struct {
	keymap     util.KeyMap
	value      *T
	renderer   func(*T) string
	editors    []editor.ValueEditor
	destructor func() tea.Cmd
	reader     func(*T) string
}

func NewStructRow[T any](
	keymap util.KeyMap,
	value *T,
	renderer func(val *T) string,
	editors []editor.ValueEditor,
) *StructRow[T] {
	return &StructRow[T]{
		keymap:   keymap,
		value:    value,
		renderer: renderer,
		editors:  editors,
	}
}

func (r *StructRow[T]) WithDestructor(callback func() tea.Cmd) *StructRow[T] {
	r.destructor = callback
	return r
}

func (r *StructRow[T]) WithReader(reader func(*T) string) *StructRow[T] {
	r.reader = reader
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
		case key.Matches(msg, r.keymap.Show) && r.reader != nil:
			return r, command.LaunchReaderScreenCmd(r.reader(r.value))
		}
	}
	return r, nil
}

func (r *StructRow[T]) View() string {
	return r.renderer(r.value)
}

func (r *StructRow[T]) Editors() []editor.ValueEditor {
	return r.editors
}

func (r *StructRow[T]) Value() *T {
	return r.value
}
