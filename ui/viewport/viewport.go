package viewport

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"hostettler.dev/dnc/util"
)

type matchPos struct {
	line int
	col  int
}

type Viewport struct {
	keymap                util.KeyMap
	cursor                int
	height                int
	width                 int
	content               []string
	highlightMatch        string
	focusedHighlightStyle lipgloss.Style
	highlightStyle        lipgloss.Style
	baseStyle             lipgloss.Style
	matches               []matchPos
	focusedMatch          int
}

func NewViewport(keymap util.KeyMap, height int, width int) *Viewport {
	return &Viewport{keymap: keymap, height: height, width: width, content: []string{}}
}

func (v *Viewport) Init() tea.Cmd {
	return nil
}

func (v *Viewport) SetHighlight(match string, focused, secondary, base lipgloss.Style) {
	v.highlightMatch = match
	v.focusedHighlightStyle = focused
	v.highlightStyle = secondary
	v.baseStyle = base
	v.findMatches()
	v.focusedMatch = 0
	v.scrollToFocused()
}

func (v *Viewport) ClearHighlight() {
	v.highlightMatch = ""
	v.matches = nil
	v.focusedMatch = 0
}

func (v *Viewport) findMatches() {
	v.matches = nil
	if v.highlightMatch == "" {
		return
	}
	for i, line := range v.content {
		col := 0
		for {
			idx := strings.Index(line[col:], v.highlightMatch)
			if idx < 0 {
				break
			}
			v.matches = append(v.matches, matchPos{line: i, col: col + idx})
			col += idx + len(v.highlightMatch)
		}
	}
}

func (v *Viewport) NextMatch() {
	if len(v.matches) == 0 {
		return
	}
	v.focusedMatch = (v.focusedMatch + 1) % len(v.matches)
	v.scrollToFocused()
}

func (v *Viewport) PrevMatch() {
	if len(v.matches) == 0 {
		return
	}
	v.focusedMatch = (v.focusedMatch - 1 + len(v.matches)) % len(v.matches)
	v.scrollToFocused()
}

func (v *Viewport) scrollToFocused() {
	if len(v.matches) == 0 {
		return
	}
	line := v.matches[v.focusedMatch].line
	if line < v.cursor {
		v.cursor = line
	} else if line >= v.cursor+v.height {
		v.cursor = line - v.height + 1
	}
	maxCursor := len(v.content) - v.height
	if v.cursor > maxCursor {
		v.cursor = maxCursor
	}
	if v.cursor < 0 {
		v.cursor = 0
	}
}

func (v *Viewport) renderLine(lineIdx int, line string) string {
	if v.highlightMatch == "" || !strings.Contains(line, v.highlightMatch) {
		return line
	}
	matchLen := len(v.highlightMatch)
	var sb strings.Builder
	pos := 0
	for i, m := range v.matches {
		if m.line < lineIdx {
			continue
		}
		if m.line > lineIdx {
			break
		}
		if m.col > pos {
			sb.WriteString(v.baseStyle.Render(line[pos:m.col]))
		}
		style := v.highlightStyle
		if i == v.focusedMatch {
			style = v.focusedHighlightStyle
		}
		sb.WriteString(style.Render(v.highlightMatch))
		pos = m.col + matchLen
	}
	if pos < len(line) {
		sb.WriteString(v.baseStyle.Render(line[pos:]))
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
		case key.Matches(msg, v.keymap.NextMatch):
			v.NextMatch()
		case key.Matches(msg, v.keymap.PrevMatch):
			v.PrevMatch()
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
		lines[i] = v.renderLine(v.cursor+i, line)
	}
	return tea.NewView(strings.Join(lines, "\n"))
}

func (v *Viewport) UpdateContent(content string) {
	bounded := lipgloss.NewStyle().Width(v.width).Render(content)
	v.content = toLines(bounded)
	v.matches = nil
	v.focusedMatch = 0
}

func (v *Viewport) Reset() {
	v.cursor = 0
	v.content = []string{}
	v.highlightMatch = ""
	v.matches = nil
	v.focusedMatch = 0
}

func toLines(s string) []string {
	return strings.Split(s, "\n")
}
