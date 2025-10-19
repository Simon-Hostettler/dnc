package editor

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/ui/util"
)

type BooleanEditor struct {
	keymap       util.KeyMap
	label        string
	value        bool
	saveCallback func(interface{}) error
	input        bool
	initialized  bool
	focus        bool
}

func NewBooleanEditor(keymap util.KeyMap, label string, delegator interface{}, saveCallback func(bool) error) *BooleanEditor {
	s := BooleanEditor{}
	fn := WrapTypedCallback(saveCallback)
	s.Init(keymap, label, delegator, fn)
	return &s
}

func (e *BooleanEditor) Init(keymap util.KeyMap, label string, delegator interface{}, saveCallback func(interface{}) error) {
	b, ok := delegator.(bool)
	if !ok {
		panic("Value passed is not a bool")
	}
	e.keymap = keymap
	e.saveCallback = saveCallback
	e.label = label
	e.value = b

	e.input = b

	e.initialized = true
}

func (e *BooleanEditor) Update(msg tea.Msg) tea.Cmd {
	if !e.initialized {
		return nil
	}

	switch m := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(m, e.keymap.Left) || key.Matches(m, e.keymap.Right):
			e.input = !e.input
		}
	}
	return nil
}

func (e *BooleanEditor) View() string {
	if !e.initialized {
		return ""
	}
	return util.RenderItem(e.focus, e.label+":") +
		" " +
		util.ItemStyleDefault.Render(util.PrettyBool(e.input))
}

func (e *BooleanEditor) Save() error {
	return e.saveCallback(e.value)
}

func (e *BooleanEditor) Focus() {
	e.focus = true
}

func (e *BooleanEditor) Blur() {
	e.focus = false
}
