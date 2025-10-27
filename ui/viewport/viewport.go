package viewport

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/util"
)

type Viewport struct {
	keymap  util.KeyMap
	cursor  int
	height  int
	width   int
	content []string
}

func NewViewport(keymap util.KeyMap, height int, width int) *Viewport {
	return &Viewport{keymap, 0, height, width, []string{}}
}

func (v *Viewport) Init() tea.Cmd {
	return nil
}

func (v *Viewport) MoveCursor(offset int) {
	newCursor := v.cursor + offset

	if newCursor >= 0 && newCursor+v.height <= len(v.content) {
		v.cursor = newCursor
	}
}

func (v *Viewport) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, v.keymap.Up):
			v.MoveCursor(-1)
		case key.Matches(msg, v.keymap.Down, v.keymap.Enter):
			v.MoveCursor(1)
		}
	}
	return v, cmd
}

func (v *Viewport) View() string {
	viewableContent := strings.Join(v.content[v.cursor:v.cursorEnd()], "\n")

	return lipgloss.NewStyle().
		Render(viewableContent)
}

func (v *Viewport) UpdateContent(content string) {
	bounded := lipgloss.NewStyle().Width(v.width).Render(content)
	v.content = toLines(bounded)
}

func (v *Viewport) Reset() {
	v.cursor = 0
	v.content = []string{}
}

func toLines(s string) []string {
	return strings.Split(s, "\n")
}

func (v *Viewport) cursorEnd() int {
	return min(len(v.content), v.cursor+v.height)
}
