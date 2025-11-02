package component

import (
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/styles"
	ui_util "hostettler.dev/dnc/util"
)

type SimpleIntComponent struct {
	keymap           ui_util.KeyMap
	name             string
	content          *int
	editor           editor.ValueEditor
	focus            bool
	renderName       bool
	highlightOnFocus bool
}

func NewSimpleIntComponent(k ui_util.KeyMap, name string, content *int, renderName bool, highlightOnFocus bool) *SimpleIntComponent {
	return &SimpleIntComponent{k, name, content, editor.NewIntEditor(k, name, content), false, renderName, highlightOnFocus}
}

func (s *SimpleIntComponent) Init() tea.Cmd {
	return nil
}

func (s *SimpleIntComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keymap.Edit):
			return s, editor.EditValueCmd([]editor.ValueEditor{s.editor})
		}
	}
	return s, nil
}

func (s *SimpleIntComponent) View() string {
	prefix := ""
	if s.renderName {
		prefix = s.name + ": "
	}
	return styles.RenderItem(s.focus && s.highlightOnFocus, prefix+strconv.Itoa(*s.content))
}

func (s *SimpleIntComponent) Focus() {
	s.focus = true
}

func (s *SimpleIntComponent) Blur() {
	s.focus = false
}

func (s *SimpleIntComponent) InFocus() bool {
	return s.focus
}
