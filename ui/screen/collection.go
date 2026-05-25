package screen

import (
	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/list"
	"hostettler.dev/dnc/util"
)

// Wires a typed slice of items to a list section: each item
// becomes a StructRow with a delete callback; a trailing AppenderRow
// adds a new item and jumps into its editor. For sectioned layouts
// (e.g. the spell screen), multiple Collections share one *list.List
// and a single onChange callback rebuilds all sections at once.
type Collection[T any] struct {
	keymap   util.KeyMap
	list     *list.List
	items    func() []*T
	idOf     func(*T) uuid.UUID
	addEmpty func() uuid.UUID
	remove   func(uuid.UUID)
	makeRow  func(*T) *list.StructRow[T]
	onChange func() // rebuilds the list; nil -> self-rebuild via Repopulate
}

func NewCollection[T any](
	keymap util.KeyMap,
	l *list.List,
	items func() []*T,
	idOf func(*T) uuid.UUID,
	addEmpty func() uuid.UUID,
	remove func(uuid.UUID),
	makeRow func(*T) *list.StructRow[T],
) *Collection[T] {
	return &Collection[T]{
		keymap: keymap, list: l, items: items, idOf: idOf,
		addEmpty: addEmpty, remove: remove, makeRow: makeRow,
	}
}

// Hook used after add/delete.
func (c *Collection[T]) WithOnChange(f func()) *Collection[T] {
	c.onChange = f
	return c
}

// Materializes the current items as a list.Section
func (c *Collection[T]) Section() list.Section {
	rows := make([]list.Row, 0, len(c.items()))
	for _, item := range c.items() {
		rows = append(rows, c.makeRow(item).WithDestructor(c.deleteCallback(c.idOf(item))))
	}
	return list.Section{
		Items:    rows,
		Appender: list.NewAppenderRow(c.keymap, c.addAndEditCmd),
	}
}

// Installs this collection as the list's sole section.
// Use only for flat (single-collection) screens; sectioned screens
// should call their own rebuild via the onChange hook instead.
func (c *Collection[T]) Repopulate() {
	c.list.WithSections([]list.Section{c.Section()})
}

// Row finds the row backing the item with the given id in the list's
// currently visible content.
func (c *Collection[T]) Row(id uuid.UUID) list.Row {
	return list.FindStructRow(c.list.Content(), func(v *T) bool {
		return c.idOf(v) == id
	})
}

func (c *Collection[T]) rebuild() {
	if c.onChange != nil {
		c.onChange()
	} else {
		c.Repopulate()
	}
}

func (c *Collection[T]) deleteCallback(id uuid.UUID) func() tea.Cmd {
	return func() tea.Cmd {
		c.remove(id)
		c.rebuild()
		return command.WriteBackRequest
	}
}

func (c *Collection[T]) addAndEditCmd() tea.Cmd {
	id := c.addEmpty()
	c.rebuild()
	row := c.Row(id)
	if row == nil {
		return nil
	}
	return editor.SwitchToEditorCmd(row.Editors())
}
