package list

import (
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/ui/editor"
	styles "hostettler.dev/dnc/ui/util"
	"hostettler.dev/dnc/util"
)

type LabeledIntRow struct {
	keymap util.KeyMap
	config LabeledIntRowConfig
	label  string
	value  *int
	editor editor.ValueEditor
}

type LabeledIntRowConfig struct {
	ValuePrinter func(int) string
	JustifyValue bool
	LabelWidth   int
	ValueWidth   int
}

func DefaultLabeledIntRowConfig() LabeledIntRowConfig {
	return LabeledIntRowConfig{strconv.Itoa, true, DefaultColWidth, DefaultColWidth}
}

func NewLabeledIntRow(keymap util.KeyMap, label string, value *int, editor editor.ValueEditor) *LabeledIntRow {
	return &LabeledIntRow{keymap, DefaultLabeledIntRowConfig(), label, value, editor}
}

func (r *LabeledIntRow) WithConfig(c LabeledIntRowConfig) *LabeledIntRow {
	r.config = c
	return r
}

func (r *LabeledIntRow) Init() tea.Cmd {
	return nil
}

func (r *LabeledIntRow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, r.keymap.Edit):
			return r, editor.EditValueCmd(r.Editors())
		}
	}
	return r, nil
}

func (r *LabeledIntRow) View() string {
	if r.config.JustifyValue {
		return styles.RenderEdgeBound(r.config.LabelWidth, r.config.ValueWidth, r.label, r.config.ValuePrinter(*r.value))
	} else {
		return styles.RenderLeftBound(r.config.LabelWidth, r.label, r.config.ValuePrinter(*r.value))
	}
}

func (r *LabeledIntRow) Editors() []editor.ValueEditor {
	return []editor.ValueEditor{r.editor}
}
