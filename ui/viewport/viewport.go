package viewport

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"hostettler.dev/dnc/util"
)

type Viewport struct {
	keymap         util.KeyMap
	cursor         int
	height         int
	width          int
	content        []string
	highlightMatch string
	highlightStyle lipgloss.Style
	baseStyle      lipgloss.Style
}

func NewViewport(keymap util.KeyMap, height int, width int) *Viewport {
	return &Viewport{keymap: keymap, height: height, width: width, content: []string{}}
}

func (v *Viewport) Init() tea.Cmd {
	return nil
}

func (v *Viewport) SetHighlight(match string, highlight lipgloss.Style, base lipgloss.Style) {
	v.highlightMatch = match
	v.highlightStyle = highlight
	v.baseStyle = base
}

func (v *Viewport) ClearHighlight() {
	v.highlightMatch = ""
}

func (v *Viewport) renderLine(line string) string {
	if v.highlightMatch == "" || !strings.Contains(line, v.highlightMatch) {
		return line
	}
	parts := strings.Split(line, v.highlightMatch)
	var sb strings.Builder
	for i, part := range parts {
		if i > 0 {
			sb.WriteString(v.highlightStyle.Render(v.highlightMatch))
		}
		if part != "" {
			sb.WriteString(v.baseStyle.Render(part))
		}
	}
	return sb.String()
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
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, v.keymap.Up):
			v.MoveCursor(-1)
		case key.Matches(msg, v.keymap.Down, v.keymap.Enter):
			v.MoveCursor(1)
		}
	}
	return v, cmd
}

func (v *Viewport) View() tea.View {
	return v.viewN(v.height)
}

func (v *Viewport) viewN(n int) tea.View {
	lines := make([]string, n)
	end := min(len(v.content), v.cursor+n)
	copy(lines, v.content[v.cursor:end])
	for i, line := range lines {
		lines[i] = v.renderLine(line)
	}
	return tea.NewView(strings.Join(lines, "\n"))
}

func (v *Viewport) UpdateContent(content string) {
	bounded := lipgloss.NewStyle().Width(v.width).Render(content)
	v.content = toLines(bounded)
}

func (v *Viewport) Reset() {
	v.cursor = 0
	v.content = []string{}
	v.highlightMatch = ""
}

func toLines(s string) []string {
	return strings.Split(s, "\n")
}
