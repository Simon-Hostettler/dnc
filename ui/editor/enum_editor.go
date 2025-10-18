package editor

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/ui/util"
)

type EnumMapping struct {
	Value int
	Label string
}

type EnumEditor struct {
	keymap       util.KeyMap
	options      []EnumMapping
	label        string
	value        int
	saveCallback func(interface{}) error
	cursor       int
	initialized  bool
	focus        bool
}

func NewEnumEditor(keymap util.KeyMap, options []EnumMapping, label string, delegator int, saveCallback func(int) error) *EnumEditor {
	e := EnumEditor{
		options: options,
	}
	fn := WrapTypedCallback(saveCallback)
	e.Init(keymap, label, delegator, fn)
	return &e
}

func (e *EnumEditor) Init(keymap util.KeyMap, label string, delegator interface{}, saveCallback func(interface{}) error) {
	e.keymap = keymap
	e.label = label
	e.saveCallback = saveCallback

	delInt, ok := delegator.(int)
	if !ok {
		panic("Value passed is not representable as an int ")
	}

	e.value = delInt

	for i, opt := range e.options {
		if opt.Value == delInt {
			e.cursor = i
			break
		}
	}

	e.initialized = true
}

func (e *EnumEditor) Update(msg tea.Msg) tea.Cmd {
	if !e.initialized {
		return nil
	}

	switch m := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(m, e.keymap.Left):
			e.cursor = (e.cursor - 1 + len(e.options)) % len(e.options)
		case key.Matches(m, e.keymap.Right, e.keymap.Select):
			e.cursor = (e.cursor + 1) % len(e.options)
		}
	}

	return nil
}

func (e *EnumEditor) View() string {
	if !e.initialized || len(e.options) == 0 {
		return ""
	}
	current := e.options[e.cursor]
	box := fmt.Sprintf("[ %s ]", current.Label)
	return util.RenderItem(e.focus, e.label+":") + " " + util.ItemStyleDefault.Render(box)
}

func (e *EnumEditor) Save() error {
	return e.saveCallback(e.value)
}

func (e *EnumEditor) Focus() {
	e.focus = true
}

func (e *EnumEditor) Blur() {
	e.focus = false
}
