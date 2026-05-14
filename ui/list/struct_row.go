package list

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/util"
)

type StructRow[T any] struct {
	id          uuid.UUID
	keymap      util.KeyMap
	value       *T
	renderer    func(*T) string
	editors     []editor.ValueEditor
	destructor  func() tea.Cmd
	reader      func(*T) string
	cycleAction func(*T) tea.Cmd
	searchText  func(*T) string
}

func NewStructRow[T any](
	keymap util.KeyMap,
	value *T,
	renderer func(val *T) string,
	editors []editor.ValueEditor,
) *StructRow[T] {
	return &StructRow[T]{
		id:       uuid.New(),
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

func (r *StructRow[T]) WithCycleAction(action func(*T) tea.Cmd) *StructRow[T] {
	r.cycleAction = action
	return r
}

// makes the row searchable
func (r *StructRow[T]) WithSearchText(searchText func(*T) string) *StructRow[T] {
	r.searchText = searchText
	return r
}

func (r *StructRow[T]) Id() uuid.UUID {
	return r.id
}

func (r *StructRow[T]) Init() tea.Cmd {
	return nil
}

func (r *StructRow[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, r.keymap.Edit):
			return r, editor.EditValueCmd(r.editors)
		case key.Matches(msg, r.keymap.Delete) && r.destructor != nil:
			return r, command.LaunchConfirmationDialogueCmd(r.destructor)
		case key.Matches(msg, r.keymap.Show) && r.reader != nil:
			return r, command.LaunchReaderScreenCmd(r.reader(r.value))
		case key.Matches(msg, r.keymap.Cycle) && r.cycleAction != nil:
			return r, r.cycleAction(r.value)
		}
	}
	return r, nil
}

func (r *StructRow[T]) View() tea.View {
	return tea.NewView(r.renderer(r.value))
}

func (r *StructRow[T]) Editors() []editor.ValueEditor {
	return r.editors
}

func (r *StructRow[T]) Value() *T {
	return r.value
}

func (r *StructRow[T]) Selectable() bool {
	return true
}

// FilterValue implements the Searchable interface. It returns an empty string
// when no search text has been configured, so the row is excluded from results
// while a search term is active.
func (r *StructRow[T]) FilterValue() string {
	if r.searchText == nil {
		return ""
	}
	return r.searchText(r.value)
}

func FindStructRow[T any](rows []Row, predicate func(*T) bool) Row {
	for _, r := range rows {
		if sr, ok := r.(*StructRow[T]); ok && predicate(sr.Value()) {
			return sr
		}
	}
	return nil
}
