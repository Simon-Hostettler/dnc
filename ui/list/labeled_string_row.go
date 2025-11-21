package list

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

type LabeledStringRow struct {
	id     uuid.UUID
	keymap util.KeyMap
	config LabeledStringRowConfig
	label  string
	value  *string
	editor *editor.StringEditor
}

type LabeledStringRowConfig struct {
	JustifyValue bool
	LabelWidth   int
	ValueWidth   int
}

func DefaultLabeledStringRowConfig() LabeledStringRowConfig {
	return LabeledStringRowConfig{true, DefaultColWidth, DefaultColWidth}
}

func NewLabeledStringRow(keymap util.KeyMap, label string, value *string, editor *editor.StringEditor) *LabeledStringRow {
	return &LabeledStringRow{uuid.New(), keymap, DefaultLabeledStringRowConfig(), label, value, editor}
}

func (r *LabeledStringRow) WithConfig(c LabeledStringRowConfig) *LabeledStringRow {
	r.config = c
	return r
}

func (r *LabeledStringRow) Id() uuid.UUID {
	return r.id
}

func (r *LabeledStringRow) Init() tea.Cmd {
	return nil
}

func (r *LabeledStringRow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, r.keymap.Edit):
			return r, editor.EditValueCmd(r.Editors())
		}
	}
	return r, nil
}

func (r *LabeledStringRow) View() string {
	if r.config.JustifyValue {
		return styles.RenderEdgeBound(r.config.LabelWidth, r.config.ValueWidth, r.label, *r.value)
	} else {
		return styles.RenderLeftBound(r.config.LabelWidth, r.label, *r.value)
	}
}

func (r *LabeledStringRow) Editors() []editor.ValueEditor {
	return []editor.ValueEditor{r.editor}
}

func (r *LabeledStringRow) Selectable() bool {
	return true
}
