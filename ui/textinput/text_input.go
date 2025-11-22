package textinput

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// lightweight wrapper that fulfills FocusableModel
type TextInput struct {
	textinput.Model
}

func New(m textinput.Model) *TextInput {
	return &TextInput{Model: m}
}

func (t *TextInput) Init() tea.Cmd { return nil }

func (t *TextInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	t.Model, cmd = t.Model.Update(msg)
	return t, cmd
}

func (t *TextInput) View() string { return t.Model.View() }

func (t *TextInput) Focus() { t.Model.Focus() }

func (t *TextInput) Blur() { t.Model.Blur() }
