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
	content string
}

func NewViewport(keymap util.KeyMap, height int, width int) *Viewport {
	return &Viewport{keymap, 0, height, width, ""}
}

func (v *Viewport) Init() tea.Cmd {
	return nil
}

func (v *Viewport) MoveCursor(offset int) {
	newCursor := v.cursor + offset

	if newCursor >= 0 && newCursor+v.height < len(v.contentToLines()) {
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
	viewableContent := strings.Join(v.contentToLines()[v.cursor:v.cursorEnd()], "\n")

	return lipgloss.NewStyle().
		MaxWidth(v.width).
		MaxHeight(v.height).
		Render(lipgloss.Place(v.width, v.height, lipgloss.Left, lipgloss.Left, viewableContent))
}

func (v *Viewport) UpdateContent(content string) {
	v.content = content
}

func (v *Viewport) Reset() {
	v.cursor = 0
	v.content = ""
}

func (v *Viewport) contentToLines() []string {
	return strings.Split(v.content, "\n")
}

func (v *Viewport) cursorEnd() int {
	return min(len(v.contentToLines()), v.cursor+v.height+1)
}
