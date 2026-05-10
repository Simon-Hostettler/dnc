package list

import (
	"strconv"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

type LabeledRow[T any] struct {
	id     uuid.UUID
	keymap util.KeyMap
	config LabeledRowConfig[T]
	label  string
	value  *T
	editor editor.ValueEditor
}

type LabeledRowConfig[T any] struct {
	ValuePrinter func(T) string
	JustifyValue bool
	LabelWidth   int
	ValueWidth   int
}

type (
	LabeledIntRowConfig    = LabeledRowConfig[int]
	LabeledStringRowConfig = LabeledRowConfig[string]
)

func newLabeledRow[T any](keymap util.KeyMap, label string, value *T, ed editor.ValueEditor, printer func(T) string) *LabeledRow[T] {
	return &LabeledRow[T]{
		id:     uuid.New(),
		keymap: keymap,
		config: LabeledRowConfig[T]{ValuePrinter: printer, JustifyValue: true, LabelWidth: DefaultColWidth, ValueWidth: DefaultColWidth},
		label:  label,
		value:  value,
		editor: ed,
	}
}

func NewLabeledIntRow(keymap util.KeyMap, label string, value *int, ed editor.ValueEditor) *LabeledRow[int] {
	return newLabeledRow(keymap, label, value, ed, strconv.Itoa)
}

func NewLabeledStringRow(keymap util.KeyMap, label string, value *string, ed editor.ValueEditor) *LabeledRow[string] {
	return newLabeledRow(keymap, label, value, ed, func(s string) string { return s })
}

func (r *LabeledRow[T]) WithConfig(c LabeledRowConfig[T]) *LabeledRow[T] {
	if c.ValuePrinter == nil {
		c.ValuePrinter = r.config.ValuePrinter
	}
	r.config = c
	return r
}

func (r *LabeledRow[T]) Id() uuid.UUID {
	return r.id
}

func (r *LabeledRow[T]) Init() tea.Cmd {
	return nil
}

func (r *LabeledRow[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, r.keymap.Edit):
			return r, editor.EditValueCmd(r.Editors())
		}
	}
	return r, nil
}

func (r *LabeledRow[T]) View() tea.View {
	if r.config.JustifyValue {
		return tea.NewView(styles.RenderEdgeBound(r.config.LabelWidth, r.config.ValueWidth, r.label, r.config.ValuePrinter(*r.value)))
	}
	return tea.NewView(styles.RenderLeftBound(r.config.LabelWidth, r.label, r.config.ValuePrinter(*r.value)))
}

func (r *LabeledRow[T]) Editors() []editor.ValueEditor {
	return []editor.ValueEditor{r.editor}
}

func (r *LabeledRow[T]) Selectable() bool {
	return true
}
