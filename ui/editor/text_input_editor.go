package editor

import (
	"log/slog"
	"strconv"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

type TextInputEditor[T any] struct {
	keymap    util.KeyMap
	label     string
	value     *T
	textInput textinput.Model
	focus     bool
	parse     func(string) (T, error)
	format    func(T) string
}

func newTextInputEditor[T any](
	keymap util.KeyMap,
	label string,
	value *T,
	parse func(string) (T, error),
	format func(T) string,
) *TextInputEditor[T] {
	ti := textinput.New()
	ti.Prompt = ""

	e := &TextInputEditor[T]{
		keymap:    keymap,
		label:     label,
		value:     value,
		textInput: ti,
		parse:     parse,
		format:    format,
	}
	e.Reload()
	return e
}

func (s *TextInputEditor[T]) Reload() {
	if s.value != nil {
		s.textInput.SetValue(s.format(*s.value))
		s.textInput.CursorEnd()
	}
}

func (s *TextInputEditor[T]) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, s.keymap.Up):
			return command.FocusNextElementCmd(command.UpDirection)
		case key.Matches(msg, s.keymap.Down):
			return command.FocusNextElementCmd(command.DownDirection)
		}
	}

	var cmd tea.Cmd
	s.textInput, cmd = s.textInput.Update(msg)
	return cmd
}

func (s *TextInputEditor[T]) View() string {
	return styles.RenderItem(s.focus, s.label+":") + " " + styles.ItemStyleDefault.Render(s.textInput.View())
}

func (s *TextInputEditor[T]) Save() tea.Cmd {
	value, err := s.parse(s.textInput.Value())
	if err != nil {
		slog.Debug("TextInputEditor: input discarded", "label", s.label, "input", s.textInput.Value(), "err", err)
		return nil
	}
	if s.value != nil {
		*s.value = value
	}
	return nil
}

func (e *TextInputEditor[T]) Focus() {
	e.textInput.Focus()
	e.focus = true
}

func (e *TextInputEditor[T]) Blur() {
	e.textInput.Blur()
	e.focus = false
}

func NewIntEditor(keymap util.KeyMap, label string, value *int) *TextInputEditor[int] {
	return newTextInputEditor(keymap, label, value, strconv.Atoi, strconv.Itoa)
}

func NewStringEditor(keymap util.KeyMap, label string, value *string) *TextInputEditor[string] {
	return newTextInputEditor(
		keymap, label, value,
		func(s string) (string, error) { return s, nil },
		func(s string) string { return s },
	)
}
