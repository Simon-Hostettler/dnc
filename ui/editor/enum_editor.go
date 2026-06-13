package editor

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

type IntLike interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type EnumEditor[T IntLike] struct {
	keymap       util.KeyMap
	options      []styles.EnumMapping
	label        string
	value        *T
	cursor       int
	focus        bool
	disabledWhen func() bool
}

func (e *EnumEditor[T]) WithDisabledWhen(predicate func() bool) *EnumEditor[T] {
	e.disabledWhen = predicate
	return e
}

func (e *EnumEditor[T]) Disabled() bool {
	return e.disabledWhen != nil && e.disabledWhen()
}

func NewEnumEditor[T IntLike](keymap util.KeyMap, options []styles.EnumMapping, label string, value *T) *EnumEditor[T] {
	e := &EnumEditor[T]{
		keymap:  keymap,
		options: options,
		label:   label,
		value:   value,
	}
	e.Reload()
	return e
}

func (e *EnumEditor[T]) Reload() {
	if e.value == nil {
		return
	}
	current := int(*e.value)
	for i, opt := range e.options {
		if opt.Value == current {
			e.cursor = i
			break
		}
	}
}

func (e *EnumEditor[T]) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	disabled := e.Disabled()
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case !disabled && key.Matches(msg, e.keymap.Left):
			e.cursor = (e.cursor - 1 + len(e.options)) % len(e.options)
		case !disabled && key.Matches(msg, e.keymap.Right, e.keymap.Select):
			e.cursor = (e.cursor + 1) % len(e.options)
		case key.Matches(msg, e.keymap.Up):
			cmd = command.FocusNextElementCmd(command.UpDirection)
		case key.Matches(msg, e.keymap.Down):
			cmd = command.FocusNextElementCmd(command.DownDirection)

		}
	}

	return cmd
}

func (e *EnumEditor[T]) View() string {
	if len(e.options) == 0 {
		return ""
	}
	current := e.options[e.cursor]
	box := fmt.Sprintf("[ %s ]", current.Label)
	if e.Disabled() {
		return styles.GrayTextStyle.Render(e.label+":") + " " + styles.GrayTextStyle.Render(box)
	}
	return styles.RenderItem(e.focus, e.label+":") + " " + styles.ItemStyleDefault.Render(box)
}

func (e *EnumEditor[T]) Save() tea.Cmd {
	if e.value == nil {
		return nil
	}
	if e.Disabled() {
		*e.value = T(e.options[0].Value)
		return nil
	}
	*e.value = T(e.options[e.cursor].Value)
	return nil
}

func (e *EnumEditor[T]) Focus() {
	e.focus = true
}

func (e *EnumEditor[T]) Blur() {
	e.focus = false
}

func (e *EnumEditor[T]) CapturesTextInput() bool {
	return false
}
