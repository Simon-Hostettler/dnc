package list

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/util"
)

var testKM = util.DefaultKeyMap()

func createDummyRow() Row {
	val := 0
	return NewLabeledIntRow(testKM, "test", &val, editor.NewIntEditor(testKM, "test", &val))
}

func TestSeparatorSkips(t *testing.T) {
	list := NewListWithDefaults(testKM).WithRows([]Row{
		createDummyRow(),
		NewSeparatorRow("-", 10),
		createDummyRow(),
	})
	list.Init()
	list.Focus()

	msg := tea.KeyPressMsg{Code: tea.KeyDown}
	list.Update(msg)

	if !(list.CursorPos() == 2) {
		t.Errorf("Separator row was not skipped, cursor at %d instead of 2", list.CursorPos())
	}
}

func TestVisibleIndexComputation(t *testing.T) {
	tmp := "tmp"
	list := NewListWithDefaults(testKM).WithRows([]Row{
		createDummyRow(),
		NewLabeledStringRow(testKM, "string", &tmp, editor.NewStringEditor(testKM, "string", &tmp)),
		createDummyRow(),
	})
	list.Init()
	list.Focus()

	list.Filter(func(r Row) bool {
		switch r.(type) {
		case *LabeledRow[string]:
			return false
		default:
			return true
		}
	})

	msg := tea.KeyPressMsg{Code: tea.KeyDown}
	list.Update(msg)

	if !(list.Size() == 2) {
		t.Errorf("Row was not filtered, size is %d instead of 2", list.Size())
	}
	if !(list.CursorPos() == 1) {
		t.Errorf("Expected cursor at 1, got %d instead", list.CursorPos())
	}
}

func TestViewPortConsistentHeight(t *testing.T) {
	list := NewListWithDefaults(testKM).WithViewport(10)
	rows := []Row{}
	for range 20 {
		rows = append(rows, createDummyRow())
	}
	list.WithRows(rows)
	list.Init()
	list.Focus()

	view := list.View().Content
	if !(lipgloss.Height(view) == 10) {
		t.Errorf("Viewport not rendering at expected height of 10. Instead %d", lipgloss.Height(view))
	}

	list.SetCursor(19)
	view = list.View().Content
	if !(lipgloss.Height(view) == 10) {
		t.Errorf("Viewport not rendering at expected height of 10. Instead %d", lipgloss.Height(view))
	}
}

func TestListExits(t *testing.T) {
	list := NewListWithDefaults(testKM).WithRows([]Row{
		createDummyRow(),
		createDummyRow(),
	})
	list.Init()
	list.Focus()

	msg := tea.KeyPressMsg{Code: tea.KeyUp}
	_, cmd := list.Update(msg)
	switch m := cmd().(type) {
	case command.FocusNextElementMsg:
		if m.Direction != command.UpDirection {
			t.Errorf("Exiting list in wrong direction. Expected: %d, Actual: %d", int(command.UpDirection), int(m.Direction))
		}
	default:
		t.Errorf("List was not exited.")
	}

	list.Focus()
	list.SetCursor(1)
	msg = tea.KeyPressMsg{Code: tea.KeyDown}
	_, cmd = list.Update(msg)
	switch m := cmd().(type) {
	case command.FocusNextElementMsg:
		if m.Direction != command.DownDirection {
			t.Errorf("Exiting list in wrong direction. Expected: %d, Actual: %d", int(command.DownDirection), int(m.Direction))
		}
	default:
		t.Errorf("List was not exited.")
	}
}

func TestRenderedRowCountWithFilter(t *testing.T) {
	tmp := "tmp"
	list := NewListWithDefaults(testKM).WithRows([]Row{
		createDummyRow(),
		NewLabeledStringRow(testKM, "string", &tmp, editor.NewStringEditor(testKM, "string", &tmp)),
		createDummyRow(),
	})
	list.Init()
	list.Focus()

	list.Filter(func(r Row) bool {
		switch r.(type) {
		case *LabeledRow[string]:
			return false
		default:
			return true
		}
	})

	view := list.View().Content
	if got, want := lipgloss.Height(view), 2; got != want {
		t.Errorf("Rendered row count mismatch. Got %d, want %d", got, want)
	}
}

func TestCursorDoesNotMoveIntoInvisibleTail(t *testing.T) {
	tmp := "tmp"
	list := NewListWithDefaults(testKM).WithRows([]Row{
		createDummyRow(),
		NewLabeledStringRow(testKM, "string", &tmp, editor.NewStringEditor(testKM, "string", &tmp)),
	})
	list.Init()
	list.Focus()

	list.Filter(func(r Row) bool {
		switch r.(type) {
		case *LabeledRow[string]:
			return false
		default:
			return true
		}
	})

	msg := tea.KeyPressMsg{Code: tea.KeyDown}
	_, cmd := list.Update(msg)

	if list.CursorPos() != 0 {
		t.Errorf("Cursor moved into/over invisible tail. Got %d, want %d", list.CursorPos(), 0)
	}
	if cmd == nil {
		t.Fatalf("expected a command, got nil")
	}
	m := cmd()
	if _, ok := m.(command.FocusNextElementMsg); !ok {
		t.Errorf("Expected FocusNextElementMsg, got %T", m)
	}
}

// --- search ---

func runeKey(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: r, Text: string(r)} }

func codeKey(c rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: c} }

func typeInto(l *List, s string) {
	for _, r := range s {
		l.Update(runeKey(r))
	}
}

// searchRow is a StructRow whose search text is its own value.
func searchRow(text string) *StructRow[string] {
	v := text
	return NewStructRow(testKM, &v, func(s *string) string { return *s }, nil).
		WithSearchText(func(s *string) string { return *s })
}

func newSearchList(rows ...Row) *List {
	l := NewList(testKM, DefaultListStyles()).WithSearch().WithRows(rows)
	l.Init()
	l.Focus()
	return l
}

func TestSearchFilter(t *testing.T) {
	fire := searchRow("Fireball")
	ice := searchRow("Ice Knife")
	sep := NewSeparatorRow("-", 10)
	app := NewAppenderRow(testKM, "x")
	// a StructRow without WithSearchText: not Searchable.
	plainVal := "plain"
	plain := NewStructRow(testKM, &plainVal, func(s *string) string { return *s }, nil)
	all := []Row{fire, ice, sep, app, plain}

	empty := SearchFilter("")
	for _, r := range all {
		if !empty(r) {
			t.Errorf("empty term should match every row, dropped %T", r)
		}
	}

	// case-insensitive, substring match against FilterValue.
	match := SearchFilter("FIRE")
	if !match(fire) {
		t.Error("'FIRE' should match 'Fireball' (case-insensitive)")
	}
	if match(ice) {
		t.Error("'FIRE' should not match 'Ice Knife'")
	}
	// rows that don't implement Searchable are dropped once a term is set.
	for _, r := range []Row{sep, app, plain} {
		if match(r) {
			t.Errorf("non-Searchable row %T must be dropped while a term is active", r)
		}
	}
}

func TestSearchOpenFilterClose(t *testing.T) {
	l := newSearchList(
		searchRow("Fireball"),
		searchRow("Fire Bolt"),
		searchRow("Ice Knife"),
	)

	// TextSearch opens and focuses the bar without filtering yet.
	l.Update(runeKey('/'))
	if !l.searchActive || !l.SearchInputFocused() {
		t.Fatal("'/' should open and focus the search bar")
	}
	if l.Size() != 3 {
		t.Errorf("opening search must not filter; got %d rows", l.Size())
	}

	// typing filters live.
	typeInto(l, "fire")
	if l.Size() != 2 {
		t.Errorf("expected 2 rows matching 'fire', got %d", l.Size())
	}

	// esc closes the bar and restores the full content.
	l.Update(codeKey(tea.KeyEscape))
	if l.searchActive {
		t.Error("esc should close the search bar")
	}
	if l.Size() != 3 {
		t.Errorf("closing search should restore all rows, got %d", l.Size())
	}
}

// The Escape binding aliases "q"; while the search input has focus, "q" must be
// typed as text rather than closing the bar.
func TestSearchInputTreatsQAsText(t *testing.T) {
	l := newSearchList(
		searchRow("Quarterstaff"),
		searchRow("Dagger"),
	)
	l.Update(runeKey('/'))

	typeInto(l, "qu")
	if !l.searchActive {
		t.Fatal("'q' must be typed into the search input, not treated as escape")
	}
	if l.searchInput.Value() != "qu" {
		t.Errorf("expected search value %q, got %q", "qu", l.searchInput.Value())
	}
	if l.Size() != 1 {
		t.Errorf("expected 1 row matching 'qu', got %d", l.Size())
	}
}

func TestSearchFocusHandoff(t *testing.T) {
	open := func() *List {
		l := newSearchList(searchRow("Aardvark"), searchRow("Badger"))
		l.Update(runeKey('/'))
		return l
	}

	t.Run("Down moves focus from the input into the rows", func(t *testing.T) {
		l := open()
		l.Update(codeKey(tea.KeyDown))
		if l.SearchInputFocused() {
			t.Error("Down should blur the search input")
		}
		if !l.searchActive {
			t.Error("Down should keep the search bar visible")
		}
		if l.CursorPos() != 0 {
			t.Errorf("Down should land on the first row, cursor at %d", l.CursorPos())
		}
	})

	t.Run("Up from the input exits the list upward", func(t *testing.T) {
		l := open()
		_, cmd := l.Update(codeKey(tea.KeyUp))
		if cmd == nil {
			t.Fatal("Up from the search input should emit a command")
		}
		msg, ok := cmd().(command.FocusNextElementMsg)
		if !ok || msg.Direction != command.UpDirection {
			t.Errorf("Up from the search input should exit the list upward, got %#v", cmd())
		}
	})

	t.Run("Up from the top row returns to the input", func(t *testing.T) {
		l := open()
		l.Update(codeKey(tea.KeyDown)) // into the rows
		l.Update(codeKey(tea.KeyUp))   // back up off the top row
		if !l.SearchInputFocused() {
			t.Error("Up from the top row should refocus the search input while searching")
		}
	})
}

// The fixed search bar must replace a body line, not grow the rendered height.
func TestSearchBarKeepsViewportHeight(t *testing.T) {
	rows := make([]Row, 0, 20)
	for range 20 {
		rows = append(rows, searchRow("row"))
	}
	l := NewList(testKM, DefaultListStyles()).
		WithFixedWidth(30).
		WithViewport(10).
		WithSearch().
		WithRows(rows)
	l.Init()
	l.Focus()

	if h := lipgloss.Height(l.View().Content); h != 10 {
		t.Fatalf("inactive search: expected viewport height 10, got %d", h)
	}

	l.Update(runeKey('/'))
	if h := lipgloss.Height(l.View().Content); h != 10 {
		t.Errorf("active search: bar must not grow total height, got %d", h)
	}
}

// Repopulating the list (e.g. after an add/delete) must keep an active filter.
func TestSearchSurvivesWithRows(t *testing.T) {
	l := newSearchList(
		searchRow("Fireball"),
		searchRow("Frostbite"),
		searchRow("Ice Knife"),
	)
	l.Update(runeKey('/'))
	typeInto(l, "fire")
	if l.Size() != 1 {
		t.Fatalf("expected 1 match for 'fire', got %d", l.Size())
	}

	l.WithRows([]Row{
		searchRow("Fireball"),
		searchRow("Fire Bolt"),
		searchRow("Frostbite"),
	})
	if l.Size() != 2 {
		t.Errorf("WithRows should re-apply the active filter; expected 2, got %d", l.Size())
	}
}
