package viewport

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

	rendered := viewport.View().Content
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

	msg := tea.KeyPressMsg{Code: tea.KeyDown}

	for range 4 {
		viewport.Update(msg)
	}

	if !(viewport.cursor == 2) {
		t.Errorf("Viewport scrolled too far down. Cursor expected at %d, is %d", 2, viewport.cursor)
	}

	msg = tea.KeyPressMsg{Code: tea.KeyUp}

	for range 4 {
		viewport.Update(msg)
	}

	if !(viewport.cursor == 0) {
		t.Errorf("Viewport scrolled too far up. Cursor expected at %d, is %d", 0, viewport.cursor)
	}
}

func setupHighlighted(v *Viewport, content, match string) {
	setupViewport(v, content)
	s := lipgloss.NewStyle()
	v.SetHighlight(match, s, s, s)
}

func TestNextMatchWrapsAround(t *testing.T) {
	viewport := NewViewport(km, 5, 10)
	setupHighlighted(viewport, "foo\nbar\nfoo\nbaz\nfoo", "foo")

	if len(viewport.matches) != 3 {
		t.Fatalf("Expected 3 matches, got %d", len(viewport.matches))
	}

	for range 3 {
		viewport.NextMatch()
	}
	if viewport.focusedMatch != 0 {
		t.Errorf("NextMatch should wrap to 0 after cycling all matches, got %d", viewport.focusedMatch)
	}
}

func TestPrevMatchWrapsFromZero(t *testing.T) {
	viewport := NewViewport(km, 5, 10)
	setupHighlighted(viewport, "foo\nbar\nfoo\nbaz\nfoo", "foo")

	viewport.PrevMatch()
	if viewport.focusedMatch != 2 {
		t.Errorf("PrevMatch from 0 should wrap to last (2), got %d", viewport.focusedMatch)
	}
}

func TestNextMatchScrollsOffScreenMatchIntoView(t *testing.T) {
	viewport := NewViewport(km, 2, 10)
	content := "foo\nx\nx\nx\nx\nfoo\nx\nx"
	setupHighlighted(viewport, content, "foo")

	if viewport.cursor != 0 {
		t.Fatalf("Initial cursor should be 0 after SetHighlight on first match, got %d", viewport.cursor)
	}

	viewport.NextMatch()

	matchLine := viewport.matches[viewport.focusedMatch].line
	if matchLine < viewport.cursor || matchLine >= viewport.cursor+viewport.height {
		t.Errorf("Focused match line %d not visible in viewport [%d, %d)",
			matchLine, viewport.cursor, viewport.cursor+viewport.height)
	}
}

func TestFindMatchesHandlesMultiplePerLine(t *testing.T) {
	viewport := NewViewport(km, 3, 40)
	setupHighlighted(viewport, "aaXaaXaa\nbb\nXX", "X")

	want := []matchPos{
		{line: 0, col: 2},
		{line: 0, col: 5},
		{line: 2, col: 0},
		{line: 2, col: 1},
	}
	if len(viewport.matches) != len(want) {
		t.Fatalf("Expected %d matches, got %d (%v)", len(want), len(viewport.matches), viewport.matches)
	}
	for i, m := range want {
		if viewport.matches[i] != m {
			t.Errorf("matches[%d] = %+v, want %+v", i, viewport.matches[i], m)
		}
	}
}

func TestNAndShiftNKeysNavigateMatches(t *testing.T) {
	viewport := NewViewport(km, 5, 10)
	setupHighlighted(viewport, "foo\nfoo\nfoo", "foo")

	viewport.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	if viewport.focusedMatch != 1 {
		t.Errorf("After 'n', focusedMatch=%d, want 1", viewport.focusedMatch)
	}

	viewport.Update(tea.KeyPressMsg{Code: 'N', Text: "N"})
	if viewport.focusedMatch != 0 {
		t.Errorf("After 'N', focusedMatch=%d, want 0", viewport.focusedMatch)
	}
}

func TestOverflowIsReadable(t *testing.T) {
	viewport := NewViewport(km, 1, 2)
	content := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	setupViewport(viewport, content)

	msg := tea.KeyPressMsg{Code: tea.KeyDown}

	var found bool
	for range len(content) / 2 {
		content, found = strings.CutPrefix(content, strings.TrimSpace(viewport.View().Content))
		if !found {
			t.Errorf("Viewport rendered something not part of the original content: %s", viewport.View().Content)
		}
		viewport.Update(msg)
	}

	if content != "" {
		t.Errorf("Viewport did not render entire content. Missing: %s", content)
	}
}
