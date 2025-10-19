package editor

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/ui/util"
)

type StringEditor struct {
	keymap       util.KeyMap
	label        string
	value        string
	saveCallback func(interface{}) error
	textInput    textinput.Model
	initialized  bool
	focus        bool
}

func NewStringEditor(keymap util.KeyMap, label string, delegator interface{}, saveCallback func(string) error) *StringEditor {
	s := StringEditor{}

	fn := WrapTypedCallback(saveCallback)
	s.Init(keymap, label, delegator, fn)
	return &s
}

func (s *StringEditor) Init(keymap util.KeyMap, label string, delegator interface{}, saveCallback func(interface{}) error) {
	str, ok := delegator.(string)
	if !ok {
		panic("Value passed is not a string")
	}
	s.value = str
	s.saveCallback = saveCallback

	ti := textinput.New()
	ti.Prompt = ""

	ti.SetValue(str)

	s.textInput = ti
	s.label = label
	s.initialized = true
}

func (s *StringEditor) Update(msg tea.Msg) tea.Cmd {
	if !s.initialized {
		return nil
	}

	var cmd tea.Cmd
	s.textInput, cmd = s.textInput.Update(msg)
	return cmd
}

func (s *StringEditor) View() string {
	if !s.initialized {
		return ""
	}
	return util.RenderItem(s.focus, s.label+":") + " " + util.ItemStyleDefault.Render(s.textInput.View())
}

func (s *StringEditor) Save() error {
	return s.saveCallback(s.value)
}

func (e *StringEditor) Focus() {
	e.textInput.Focus()
	e.focus = true
}

func (e *StringEditor) Blur() {
	e.textInput.Blur()
	e.focus = false
}
