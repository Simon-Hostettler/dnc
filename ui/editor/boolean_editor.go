package editor

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	styles "hostettler.dev/dnc/ui/util"
	"hostettler.dev/dnc/util"
)

type BooleanEditor struct {
	keymap      util.KeyMap
	label       string
	value       *bool
	input       bool
	initialized bool
	focus       bool
}

func NewBooleanEditor(keymap util.KeyMap, label string, delegatorPointer interface{}) *BooleanEditor {
	s := BooleanEditor{}
	s.Init(keymap, label, delegatorPointer)
	return &s
}

func (e *BooleanEditor) Init(keymap util.KeyMap, label string, delegatorPointer interface{}) {
	ptr, ok := delegatorPointer.(*bool)
	if !ok {
		panic("Value passed is not a pointer to bool")
	}
	e.keymap = keymap
	e.label = label
	e.value = ptr

	if ptr != nil {
		e.input = *ptr
	}

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
	return styles.RenderItem(e.focus, e.label+":") +
		" " +
		styles.ItemStyleDefault.Render(styles.PrettyBool(e.input))
}

func (e *BooleanEditor) Save() tea.Cmd {
	if e.value != nil {
		*e.value = e.input
	}
	return nil
}

func (e *BooleanEditor) Focus() {
	e.focus = true
}

func (e *BooleanEditor) Blur() {
	e.focus = false
}
