package screen

import (
	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/list"
	"hostettler.dev/dnc/util"
)

/*
Centralizes the append / delete / repopulate boilerplate for list-backed screen sections
*/
type CollectionRows[T any] struct {
	keymap   util.KeyMap
	list     *list.List
	tag      string
	items    func() []*T
	idOf     func(*T) uuid.UUID
	addEmpty func(tag string) uuid.UUID
	remove   func(uuid.UUID)
	makeRow  func(*T) *list.StructRow[T]

	Repopulate func()
}

// NewCollectionRows wires a collection whose list is a flat row-per-item
// layout with a single trailing appender.
func NewCollectionRows[T any](
	keymap util.KeyMap,
	l *list.List,
	tag string,
	items func() []*T,
	idOf func(*T) uuid.UUID,
	addEmpty func() uuid.UUID,
	remove func(uuid.UUID),
	makeRow func(*T) *list.StructRow[T],
) *CollectionRows[T] {
	c := &CollectionRows[T]{
		keymap:   keymap,
		list:     l,
		tag:      tag,
		items:    items,
		idOf:     idOf,
		addEmpty: func(string) uuid.UUID { return addEmpty() },
		remove:   remove,
		makeRow:  makeRow,
	}
	c.Repopulate = c.flatRepopulate
	return c
}

// NewCustomCollectionRows wires a collection whose row layout is built by the
// caller. The caller must assign Repopulate before use. addEmpty receives the
// AppenderRow tag so layouts with multiple appenders can disambiguate.
func NewCustomCollectionRows[T any](
	l *list.List,
	idOf func(*T) uuid.UUID,
	addEmpty func(tag string) uuid.UUID,
	remove func(uuid.UUID),
) *CollectionRows[T] {
	return &CollectionRows[T]{
		list:     l,
		idOf:     idOf,
		addEmpty: addEmpty,
		remove:   remove,
	}
}

func (c *CollectionRows[T]) flatRepopulate() {
	rows := []list.Row{}
	for _, item := range c.items() {
		rows = append(rows, c.makeRow(item).WithDestructor(c.DeleteCallback(c.idOf(item))))
	}
	rows = append(rows, list.NewAppenderRow(c.keymap, c.tag))
	c.list.WithRows(rows)
}

// Row finds the row backing the item with the given id.
func (c *CollectionRows[T]) Row(id uuid.UUID) list.Row {
	return list.FindStructRow(c.list.Content(), func(v *T) bool {
		return c.idOf(v) == id
	})
}

func (c *CollectionRows[T]) DeleteCallback(id uuid.UUID) func() tea.Cmd {
	return func() tea.Cmd {
		c.remove(id)
		c.Repopulate()
		return command.WriteBackRequest
	}
}

// Adds an empty item, rebuilds the rows, and jumps straight into editor.
// Returned from a screen's command.AppendElementMsg handler.
func (c *CollectionRows[T]) HandleAppend(tag string) tea.Cmd {
	id := c.addEmpty(tag)
	c.Repopulate()
	return editor.SwitchToEditorCmd(c.Row(id).Editors())
}
