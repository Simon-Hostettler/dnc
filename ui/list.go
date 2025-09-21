package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Row interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
	Editors() []ValueEditor
}

type ListStyles struct {
	Row      lipgloss.Style
	Selected lipgloss.Style
}

func DefaultListStyles() ListStyles {
	return ListStyles{
		Row:      ItemStyleDefault,
		Selected: ItemStyleSelected,
	}
}

type List struct {
	KeyMap KeyMap
	Styles ListStyles

	focus      bool
	title      string
	content    []Row
	cursor     int
	appendable bool
}

func (t *List) WithKeyMap(k KeyMap) *List {
	t.KeyMap = k
	return t
}

func (t *List) WithStyles(s ListStyles) *List {
	t.Styles = s
	return t
}

func (t *List) WithRows(r []Row) *List {
	t.content = r
	return t
}

func (t *List) WithTitle(title string) *List {
	t.title = title
	return t
}

/*
The appender will simply send out an AppendElementCmd,
the implementation is the client's responsibility.
*/
func (t *List) WithAppender() *List {
	t.appendable = true
	return t
}

func (t *List) Focus() {
	t.focus = true
}

func (t *List) Blur() {
	t.focus = false
}

func (t *List) InFocus() bool {
	return t.focus
}

func (t *List) SetCursor(idx int) {
	if !(idx < 0 || idx > len(t.content)) {
		t.cursor = idx
	}
}

func (t *List) MoveCursor(offset int) tea.Cmd {
	newCursor := t.cursor + offset
	if newCursor >= 0 && newCursor < len(t.content)+B2i(t.appendable) {
		t.cursor = newCursor
		return nil
	} else {
		if newCursor < 0 {
			return FocusNextElementCmd(UpDirection)
		} else {
			return FocusNextElementCmd(DownDirection)
		}
	}
}

func NewList(k KeyMap, s ListStyles) *List {
	return &List{
		KeyMap: k,
		Styles: s,
	}
}

func NewListWithDefaults() *List {
	return &List{
		KeyMap: DefaultKeyMap(),
		Styles: DefaultListStyles(),
	}
}

func (t *List) Init() tea.Cmd {
	return nil
}

func (t *List) Update(m tea.Msg) (tea.Model, tea.Cmd) {
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
		case key.Matches(msg, t.KeyMap.Select) && t.cursor == len(t.content):
			cmd = AppendElementCmd
		default:
			_, cmd = t.content[t.cursor].Update(m)
		}
	}
	return t, cmd
}

func (t *List) View() string {
	body := t.RenderBody()
	if t.title != "" {
		var title string
		if t.focus {
			title = t.Styles.Selected.Render(t.title) + "\n"
		} else {
			title = t.Styles.Row.Render(t.title) + "\n"
		}
		body = lipgloss.JoinVertical(lipgloss.Center, title, body)
	}
	return body
}

func (t *List) RenderBody() string {
	rows := []string{}

	for i, el := range t.content {
		elStr := el.View()
		if t.focus && i == t.cursor {
			rows = append(rows,
				t.Styles.Selected.Render(elStr))
		} else {
			rows = append(rows,
				t.Styles.Row.Render(elStr))
		}
	}
	list := lipgloss.JoinVertical(lipgloss.Left, rows...)
	if t.appendable {
		var adder string
		if t.focus && t.cursor == len(t.content) {
			adder = "\n" + t.Styles.Selected.Render("+")
		} else {
			adder = "\n" + t.Styles.Row.Render("+")
		}
		return lipgloss.JoinVertical(lipgloss.Center, list, adder)
	} else {
		return list
	}
}
