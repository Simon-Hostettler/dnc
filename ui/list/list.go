package list

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/ui/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/util"
)

var DefaultColWidth = 16

type Row interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
	Editors() []editor.ValueEditor
}

type ListStyles struct {
	Row      lipgloss.Style
	Selected lipgloss.Style
}

func DefaultListStyles() ListStyles {
	return ListStyles{
		Row:      util.ItemStyleDefault,
		Selected: util.ItemStyleSelected,
	}
}

type List struct {
	KeyMap util.KeyMap
	Styles ListStyles

	focus      bool
	title      string
	content    []Row
	cursor     int
	fixedWidth int
}

func (t *List) WithKeyMap(k util.KeyMap) *List {
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

func (t *List) WithFixedWidth(width int) *List {
	t.fixedWidth = width
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

func (t *List) Size() int {
	return len(t.content)
}

func (t *List) Content() []Row {
	return t.content
}

func (t *List) CursorPos() int {
	return t.cursor
}

func (t *List) SetCursor(idx int) {
	if !(idx < 0 || idx > len(t.content)) {
		t.cursor = idx
	}
}

func (t *List) MoveCursor(offset int) tea.Cmd {
	newCursor := t.cursor + offset
	if newCursor >= 0 && newCursor < len(t.content) {
		t.cursor = newCursor
		switch t.content[newCursor].(type) {
		case *SeparatorRow: // not selectable, skip over
			return t.MoveCursor(offset)
		default:
			return nil
		}
	} else {
		if newCursor < 0 {
			return command.FocusNextElementCmd(command.UpDirection)
		} else {
			return command.FocusNextElementCmd(command.DownDirection)
		}
	}
}

func NewList(k util.KeyMap, s ListStyles) *List {
	return &List{
		KeyMap:     k,
		Styles:     s,
		fixedWidth: -1,
	}
}

func NewListWithDefaults() *List {
	return &List{
		KeyMap: util.DefaultKeyMap(),
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
		default:
			if t.cursor < len(t.content) {
				_, cmd = t.content[t.cursor].Update(m)
			}
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
		var row string
		if t.focus && i == t.cursor {
			if t.fixedWidth != -1 {
				row = t.Styles.Selected.Width(t.fixedWidth).Render(elStr)
			} else {
				row = t.Styles.Selected.Render(elStr)
			}
		} else {
			if t.fixedWidth != -1 {
				row = t.Styles.Row.Width(t.fixedWidth).Render(elStr)
			} else {
				row = t.Styles.Row.Render(elStr)
			}
		}
		rows = append(rows, row)
	}
	list := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return list
}
