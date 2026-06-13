package editor

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

type TextEditor struct {
	keymap   util.KeyMap
	label    string
	value    *string
	textArea textarea.Model
	focus    bool
}

func NewTextEditor(keymap util.KeyMap, label string, value *string) *TextEditor {
	ta := textarea.New()
	ta.SetWidth(styles.SmallScreenWidth - 4)
	ta.SetHeight(8)
	ta.ShowLineNumbers = false
	ta.Prompt = ""

	e := &TextEditor{
		keymap:   keymap,
		label:    label,
		value:    value,
		textArea: ta,
	}
	e.Reload()
	return e
}

func (e *TextEditor) Reload() {
	if e.value != nil {
		e.textArea.SetValue(*e.value)
	}
	e.textArea.CursorStart()
}

func (e *TextEditor) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, e.keymap.Up) &&
			e.textArea.Line() == 0 &&
			e.textArea.LineInfo().RowOffset == 0:
			return command.FocusNextElementCmd(command.UpDirection)
		case key.Matches(msg, e.keymap.Down) &&
			e.textArea.Line() == e.textArea.LineCount()-1 &&
			e.textArea.LineInfo().RowOffset == e.textArea.LineInfo().Height-1:
			return command.FocusNextElementCmd(command.DownDirection)
		}
	}

	var cmd tea.Cmd
	e.textArea, cmd = e.textArea.Update(msg)
	return cmd
}

func (e *TextEditor) View() string {
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

func (e *TextEditor) CapturesTextInput() bool {
	return true
}
