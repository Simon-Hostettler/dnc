package list

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/ui/command"
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

	msg := tea.KeyMsg{Type: tea.KeyDown}
	list.Update(msg)
	list.Update(msg)

	if !(list.CursorPos() == 2) {
		t.Errorf("Separator row was not skipped, cursor at %d instead of 2", list.CursorPos())
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

	view := list.View()
	if !(lipgloss.Height(view) == 10) {
		t.Errorf("Viewport not rendering at expected height of 10. Instead %d", lipgloss.Height(view))
	}

	list.SetCursor(19)
	view = list.View()
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

	msg := tea.KeyMsg{Type: tea.KeyUp}
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
	msg = tea.KeyMsg{Type: tea.KeyDown}
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
