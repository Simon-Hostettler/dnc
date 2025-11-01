package editor

import (
	"fmt"
	"reflect"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui/command"
	styles "hostettler.dev/dnc/ui/util"
	"hostettler.dev/dnc/util"
)

type EnumEditor struct {
	keymap      util.KeyMap
	options     []models.EnumMapping
	label       string
	value       reflect.Value
	cursor      int
	initialized bool
	focus       bool
}

func NewEnumEditor(keymap util.KeyMap, options []models.EnumMapping, label string, delegatorPointer interface{}) *EnumEditor {
	e := EnumEditor{
		options: options,
	}
	e.Init(keymap, label, delegatorPointer)
	return &e
}

func (e *EnumEditor) Init(keymap util.KeyMap, label string, delegatorPointer interface{}) {
	e.keymap = keymap

	ptrValue := reflect.ValueOf(delegatorPointer)
	if ptrValue.Kind() != reflect.Ptr || !ptrValue.Elem().IsValid() {
		panic("Value passed is not a valid pointer")
	}

	elem := ptrValue.Elem()
	kind := elem.Kind()
	if kind < reflect.Int || kind > reflect.Int64 {
		panic(fmt.Sprintf("Value passed is not a pointer to int-like, got: %s", kind))
	}

	e.value = ptrValue

	currentValue := int(elem.Int())
	for i, opt := range e.options {
		if opt.Value == currentValue {
			e.cursor = i
			break
		}
	}

	e.label = label
	e.initialized = true
}

func (e *EnumEditor) Update(msg tea.Msg) tea.Cmd {
	if !e.initialized {
		return nil
	}

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, e.keymap.Left):
			e.cursor = (e.cursor - 1 + len(e.options)) % len(e.options)
		case key.Matches(msg, e.keymap.Right, e.keymap.Select):
			e.cursor = (e.cursor + 1) % len(e.options)
		case key.Matches(msg, e.keymap.Up):
			cmd = command.FocusNextElementCmd(command.UpDirection)
		case key.Matches(msg, e.keymap.Down):
			cmd = command.FocusNextElementCmd(command.DownDirection)

		}
	}

	return cmd
}

func (e *EnumEditor) View() string {
	if !e.initialized || len(e.options) == 0 {
		return ""
	}
	current := e.options[e.cursor]
	box := fmt.Sprintf("[ %s ]", current.Label)
	return styles.RenderItem(e.focus, e.label+":") + " " + styles.ItemStyleDefault.Render(box)
}

func (e *EnumEditor) Save() tea.Cmd {
	e.value.Elem().SetInt(int64(e.options[e.cursor].Value))
	return nil
}

func (e *EnumEditor) Focus() {
	e.focus = true
}

func (e *EnumEditor) Blur() {
	e.focus = false
}
