package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type LabeledStringRow struct {
	keymap KeyMap
	config LabeledStringRowConfig
	label  string
	value  *string
	editor *StringEditor
}

type LabeledStringRowConfig struct {
	JustifyValue bool
	LabelWidth   int
	ValueWidth   int
}

func DefaultLabeledStringRowConfig() LabeledStringRowConfig {
	return LabeledStringRowConfig{true, ColWidth, ColWidth}
}

func NewLabeledStringRow(keymap KeyMap, label string, value *string, editor *StringEditor) *LabeledStringRow {
	return &LabeledStringRow{keymap, DefaultLabeledStringRowConfig(), label, value, editor}
}

func (r *LabeledStringRow) WithConfig(c LabeledStringRowConfig) *LabeledStringRow {
	r.config = c
	return r
}

func (r *LabeledStringRow) Init() tea.Cmd {
	return nil
}

func (r *LabeledStringRow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, r.keymap.Edit):
			return r, EditValueCmd(r.Editors())
		}

	}
	return r, nil
}

func (r *LabeledStringRow) View() string {
	if r.config.JustifyValue {
		return RenderEdgeBound(r.config.LabelWidth, r.config.ValueWidth, r.label, *r.value)
	} else {
		return RenderLeftBound(r.config.LabelWidth, r.label, *r.value)
	}
}

func (r *LabeledStringRow) Editors() []ValueEditor {
	return []ValueEditor{r.editor}
}
