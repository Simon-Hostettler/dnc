package viewport

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/util"
)

var km = util.DefaultKeyMap()

func TestRespectsDimensions(t *testing.T) {
	viewport := NewViewport(km, 1, 5)
	content := "ABCDEFG\n" + "HIJKLMN"

	cmd := viewport.Init()
	var m tea.Msg
	if cmd != nil {
		m = cmd()
	}
	viewport.Update(m)

	viewport.UpdateContent(content)

	rendered := viewport.View()
	if !(lipgloss.Height(rendered) == 1) {
		t.Errorf("Viewport rendered with unexpected height: %d", lipgloss.Height(rendered))
	}
	if !(lipgloss.Width(rendered) == 5) {
		t.Errorf("Viewport rendered with unexpected width: %d", lipgloss.Width(rendered))
	}
}

func TestCursorStaysInBounds(t *testing.T) {
	viewport := NewViewport(km, 3, 5)
	content := ""
	for range 5 {
		content += "ABCDE\n"
	}

	cmd := viewport.Init()
	var m tea.Msg
	if cmd != nil {
		m = cmd()
	}
	viewport.Update(m)

	viewport.UpdateContent(content)

	msg := tea.KeyMsg{Type: tea.KeyDown}

	for range 4 {
		viewport.Update(msg)
	}

	if !(viewport.cursor == 3) {
		t.Errorf("Viewport scrolled too far down. Cursor expected at %d, is %d", 3, viewport.cursor)
	}

	msg = tea.KeyMsg{Type: tea.KeyUp}

	for range 4 {
		viewport.Update(msg)
	}

	if !(viewport.cursor == 0) {
		t.Errorf("Viewport scrolled too far up. Cursor expected at %d, is %d", 0, viewport.cursor)
	}
}
