package editor

import (
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/ui/util"
)

type IntEditor struct {
	keymap       util.KeyMap
	label        string
	value        int
	saveCallback func(interface{}) error
	textInput    textinput.Model
	initialized  bool
	focus        bool
}

func NewIntEditor(keymap util.KeyMap, label string, delegator int, saveCallback func(int) error) *IntEditor {
	s := IntEditor{}

	fn := WrapTypedCallback(saveCallback)
	s.Init(keymap, label, delegator, fn)
	return &s
}

func (s *IntEditor) Init(keymap util.KeyMap, label string, delegator interface{}, saveCallback func(interface{}) error) {
	delInt, ok := delegator.(int)
	if !ok {
		panic("Value passed is not an int")
	}
	s.value = delInt
	s.saveCallback = saveCallback

	ti := textinput.New()
	ti.Prompt = ""

	ti.SetValue(strconv.Itoa(delInt))

	s.textInput = ti
	s.label = label
	s.initialized = true
}

func (s *IntEditor) Update(msg tea.Msg) tea.Cmd {
	if !s.initialized {
		return nil
	}

	var cmd tea.Cmd
	s.textInput, cmd = s.textInput.Update(msg)
	return cmd
}

func (s *IntEditor) View() string {
	if !s.initialized {
		return ""
	}
	return util.RenderItem(s.focus, s.label+":") + " " + util.ItemStyleDefault.Render(s.textInput.View())
}

func (s *IntEditor) Save() error {
	value, err := strconv.Atoi(s.textInput.Value())
	if err != nil {
		return err
	}
	return s.saveCallback(value)
}

func (e *IntEditor) Focus() {
	e.textInput.Focus()
	e.focus = true
}

func (e *IntEditor) Blur() {
	e.textInput.Blur()
	e.focus = false
}
