package editor

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

type TextEditor struct {
	keymap      util.KeyMap
	label       string
	value       *string
	textArea    textarea.Model
	initialized bool
	focus       bool
}

func NewTextEditor(keymap util.KeyMap, label string, delegatorPointer interface{}) *TextEditor {
	s := TextEditor{}
	s.Init(keymap, label, delegatorPointer)
	return &s
}

func (e *TextEditor) Init(keymap util.KeyMap, label string, delegatorPointer interface{}) {
	ptr, ok := delegatorPointer.(*string)
	if !ok {
		panic("Value passed is not a pointer to string")
	}
	e.keymap = keymap
	e.value = ptr

	ta := textarea.New()
	ta.SetWidth(styles.SmallScreenWidth - 4)
	ta.ShowLineNumbers = false
	ta.Prompt = ""

	if ptr != nil {
		ta.SetValue(*ptr)
	}

	e.textArea = ta
	e.label = label
	e.initialized = true
}

func (e *TextEditor) Update(msg tea.Msg) tea.Cmd {
	if !e.initialized {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, e.keymap.Up) && e.textArea.Line() == 0:
			return command.FocusNextElementCmd(command.UpDirection)
		case key.Matches(msg, e.keymap.Down) && e.textArea.Line() == e.textArea.LineCount()-1:
			return command.FocusNextElementCmd(command.DownDirection)
		}
	}

	var cmd tea.Cmd
	e.textArea, cmd = e.textArea.Update(msg)
	return cmd
}

func (e *TextEditor) View() string {
	if !e.initialized {
		return ""
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		styles.RenderItem(e.focus, e.label+":"),
		styles.ItemStyleDefault.Render(e.textArea.View()),
	)
}

func (e *TextEditor) Save() tea.Cmd {
	if e.value != nil {
		*e.value = e.textArea.Value()
	}
	return nil
}

func (e *TextEditor) Focus() {
	e.textArea.Focus()
	e.focus = true
}

func (e *TextEditor) Blur() {
	e.textArea.Blur()
	e.focus = false
}
