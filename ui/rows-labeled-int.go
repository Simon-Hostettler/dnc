package ui

import (
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type LabeledIntRow struct {
	keymap KeyMap
	config LabeledIntRowConfig
	label  string
	value  *int
	editor ValueEditor
}

type LabeledIntRowConfig struct {
	ValuePrinter func(int) string
	JustifyValue bool
	LabelWidth   int
	ValueWidth   int
}

func DefaultLabeledIntRowConfig() LabeledIntRowConfig {
	return LabeledIntRowConfig{strconv.Itoa, true, ColWidth, ColWidth}
}

func NewLabeledIntRow(keymap KeyMap, label string, value *int, editor ValueEditor) *LabeledIntRow {
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
			return r, EditValueCmd(r.Editors())
		}

	}
	return r, nil
}

func (r *LabeledIntRow) View() string {
	if r.config.JustifyValue {
		return RenderEdgeBound(r.config.LabelWidth, r.config.ValueWidth, r.label, r.config.ValuePrinter(*r.value))
	} else {
		return RenderLeftBound(r.config.LabelWidth, r.label, r.config.ValuePrinter(*r.value))
	}
}

func (r *LabeledIntRow) Editors() []ValueEditor {
	return []ValueEditor{r.editor}
}
