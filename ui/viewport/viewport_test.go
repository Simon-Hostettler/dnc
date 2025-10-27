package viewport

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/util"
)

var km = util.DefaultKeyMap()

func setupViewport(v *Viewport, content string) {
	cmd := v.Init()
	var m tea.Msg
	if cmd != nil {
		m = cmd()
	}
	v.Update(m)

	v.UpdateContent(content)
}

func TestRespectsDimensions(t *testing.T) {
	viewport := NewViewport(km, 1, 5)
	content := "ABCDEFG\n" + "HIJKLMN"

	setupViewport(viewport, content)

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
	content := "ABCDE"
	for range 4 {
		content += "\nABCDE"
	}

	setupViewport(viewport, content)

	msg := tea.KeyMsg{Type: tea.KeyDown}

	for range 4 {
		viewport.Update(msg)
	}

	if !(viewport.cursor == 2) {
		t.Errorf("Viewport scrolled too far down. Cursor expected at %d, is %d", 2, viewport.cursor)
	}

	msg = tea.KeyMsg{Type: tea.KeyUp}

	for range 4 {
		viewport.Update(msg)
	}

	if !(viewport.cursor == 0) {
		t.Errorf("Viewport scrolled too far up. Cursor expected at %d, is %d", 0, viewport.cursor)
	}
}

func TestOverflowIsReadable(t *testing.T) {
	viewport := NewViewport(km, 1, 2)
	content := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	setupViewport(viewport, content)

	msg := tea.KeyMsg{Type: tea.KeyDown}

	var found bool
	for range len(content) / 2 {
		content, found = strings.CutPrefix(content, strings.TrimSpace(viewport.View()))
		if !found {
			t.Errorf("Viewport rendered something not part of the original content: %s", viewport.View())
		}
		viewport.Update(msg)
	}

	if content != "" {
		t.Errorf("Viewport did not render entire content. Missing: %s", content)
	}
}
