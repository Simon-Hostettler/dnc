package editor

import (
	"fmt"
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
	keymap      util.KeyMap
	label       string
	value       *T
	textInput   textinput.Model
	initialized bool
	focus       bool
	parse       func(string) (T, error)
	format      func(T) string
}

func (s *TextInputEditor[T]) Init(keymap util.KeyMap, label string, delegatorPointer interface{}) {
	ptr, ok := delegatorPointer.(*T)
	if !ok {
		var zero T
		panic(fmt.Sprintf("Value passed is not a pointer to %T", zero))
	}
	s.keymap = keymap
	s.value = ptr

	ti := textinput.New()
	ti.Prompt = ""

	if ptr != nil {
		ti.SetValue(s.format(*ptr))
	}

	s.textInput = ti
	s.label = label
	s.initialized = true
}

func (s *TextInputEditor[T]) Update(msg tea.Msg) tea.Cmd {
	if !s.initialized {
		return nil
	}

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
	if !s.initialized {
		return ""
	}
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

func NewIntEditor(keymap util.KeyMap, label string, delegatorPointer interface{}) *TextInputEditor[int] {
	e := &TextInputEditor[int]{
		parse:  strconv.Atoi,
		format: strconv.Itoa,
	}
	e.Init(keymap, label, delegatorPointer)
	return e
}

func NewStringEditor(keymap util.KeyMap, label string, delegatorPointer interface{}) *TextInputEditor[string] {
	e := &TextInputEditor[string]{
		parse:  func(s string) (string, error) { return s, nil },
		format: func(s string) string { return s },
	}
	e.Init(keymap, label, delegatorPointer)
	return e
}
