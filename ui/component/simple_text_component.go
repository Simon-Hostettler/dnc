package component

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

type SimpleTextComponent struct {
	keymap  util.KeyMap
	name    string
	content *string
	editor  editor.ValueEditor
	focus   bool
	height  int
	width   int
}

func NewSimpleTextComponent(k util.KeyMap, name string, content *string, height int, width int) *SimpleTextComponent {
	return &SimpleTextComponent{k, name, content, editor.NewTextEditor(k, name, content), false, height, width}
}

func (s *SimpleTextComponent) Init() tea.Cmd {
	return nil
}

func (s *SimpleTextComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keymap.Edit):
			return s, editor.EditValueCmd([]editor.ValueEditor{s.editor})
		case key.Matches(msg, s.keymap.Show):
			return s, command.LaunchReaderScreenCmd(*s.content)
		}
	}
	return s, nil
}

func (s *SimpleTextComponent) View() string {
	return styles.RenderTextBox(*s.content, s.width, s.height)
}

func (s *SimpleTextComponent) Focus() {
	s.focus = true
}

func (s *SimpleTextComponent) Blur() {
	s.focus = false
}

func (s *SimpleTextComponent) InFocus() bool {
	return s.focus
}
