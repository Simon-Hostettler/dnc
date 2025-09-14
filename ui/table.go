package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Row []string

type TableStyles struct {
	Title    lipgloss.Style
	Row      lipgloss.Style
	Selected lipgloss.Style
}

func DefaultTableStyles() TableStyles {
	return TableStyles{
		Title:    DefaultTextStyle,
		Row:      ItemStyleDefault,
		Selected: ItemStyleSelected,
	}
}

type Table struct {
	KeyMap KeyMap
	Styles TableStyles

	rowHandler func(KeyMap, tea.Msg, Row) tea.Cmd
	focus      bool
	title      string
	content    []Row
	cursor     int
}

func (t *Table) WithKeyMap(k KeyMap) *Table {
	t.KeyMap = k
	return t
}

func (t *Table) WithStyles(s TableStyles) *Table {
	t.Styles = s
	return t
}

func (t *Table) WithRows(r []Row) *Table {
	t.content = r
	return t
}

func (t *Table) WithTitle(title string) *Table {
	t.title = title
	return t
}

func (t *Table) WithRowHandler(f func(KeyMap, tea.Msg, Row) tea.Cmd) *Table {
	t.rowHandler = f
	return t
}

func (t *Table) SetFocus(f bool) *Table {
	t.focus = f
	return t
}

func (t *Table) IsFocus() bool {
	return t.focus
}

func (t *Table) MoveCursor(offset int) tea.Cmd {
	newCursor := t.cursor + offset
	if newCursor >= 0 && newCursor < len(t.content) {
		t.cursor = newCursor
		return nil
	} else {
		if newCursor < 0 {
			return ExitTableCmd(tea.KeyUp)
		} else {
			return ExitTableCmd(tea.KeyDown)
		}

	}
}

func NewTable(k KeyMap, s TableStyles) *Table {
	return &Table{
		KeyMap: k,
		Styles: s,
	}
}

func NewTableWithDefaults() *Table {
	return &Table{
		KeyMap: DefaultKeyMap(),
		Styles: DefaultTableStyles(),
	}
}

func (t *Table) Init() tea.Cmd {
	return nil
}

func (t *Table) Update(m tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if !t.focus {
		return t, nil
	}

	switch msg := m.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, t.KeyMap.Up):
			cmd = t.MoveCursor(-1)
		case key.Matches(msg, t.KeyMap.Down):
			cmd = t.MoveCursor(1)
		case key.Matches(msg, t.KeyMap.Escape):
			t.focus = false
		default:
			cmd = t.rowHandler(t.KeyMap, m, t.content[t.cursor])
		}
	}
	return t, cmd
}

func (t *Table) View() string {
	body := t.RenderBody()
	if t.title != "" {
		title := t.Styles.Title.Render(t.title) + "\n"
		body = lipgloss.JoinVertical(lipgloss.Center, title, body)
	}
	return body
}

func (t *Table) RenderBody() string {
	s := ""

	for i, el := range t.content {
		elStr := lipgloss.JoinHorizontal(lipgloss.Left, el...)
		if t.focus && i == t.cursor {
			s += t.Styles.Selected.Render(elStr) + "\n"
		} else {
			s += t.Styles.Row.Render(elStr) + "\n"
		}
	}
	return s
}
