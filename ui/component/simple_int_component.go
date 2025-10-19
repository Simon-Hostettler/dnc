package component

import (
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/util"
	ui_util "hostettler.dev/dnc/ui/util"
)

type SimpleIntComponent struct {
	keymap           ui_util.KeyMap
	name             string
	value            int
	editor           editor.ValueEditor
	focus            bool
	renderName       bool
	highlightOnFocus bool
}

func NewSimpleIntComponent(k ui_util.KeyMap, name string, value int, saveCallback func(int) error, renderName bool, highlightOnFocus bool) *SimpleIntComponent {
	return &SimpleIntComponent{k, name, value, editor.NewIntEditor(k, name, value, saveCallback), false, renderName, highlightOnFocus}
}

func (s *SimpleIntComponent) WithSaveCallback(saveCallback func(int) error) *SimpleIntComponent {
	s.editor = editor.NewIntEditor(s.keymap, s.name, s.value, saveCallback)
	return s
}

func (s *SimpleIntComponent) WithValue(v int) *SimpleIntComponent {
	s.value = v
	return s
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
	return util.RenderItem(s.focus && s.highlightOnFocus, prefix+strconv.Itoa(s.value))
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
