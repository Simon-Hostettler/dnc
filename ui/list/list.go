package list

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/google/uuid"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

var DefaultColWidth = 16

const searchBarHeight = 1

// closes the search bar from within the input. Deliberately matches only the literal escape key
// so all other keys remain typable
var searchEscapeKey = key.NewBinding(key.WithKeys("esc"))

type Row interface {
	Id() uuid.UUID
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() tea.View
	Editors() []editor.ValueEditor
	Selectable() bool
}

// Implemented by rows that can be matched against a search term.
type Searchable interface {
	FilterValue() string
}

// case-insensitive, empty -> all match
func SearchFilter(term string) func(Row) bool {
	normalized := strings.ToLower(strings.TrimSpace(term))
	if normalized == "" {
		return func(Row) bool { return true }
	}
	return func(r Row) bool {
		s, ok := r.(Searchable)
		return ok && strings.Contains(strings.ToLower(s.FilterValue()), normalized)
	}
}

type ListStyles struct {
	Row      lipgloss.Style
	Selected lipgloss.Style
}

func DefaultListStyles() ListStyles {
	return ListStyles{
		Row:      styles.ItemStyleDefault,
		Selected: styles.ItemStyleSelected,
	}
}

type List struct {
	KeyMap util.KeyMap
	Styles ListStyles

	focus      bool
	title      string
	content    []Row
	visible    []Row
	cursor     int
	fixedWidth int

	viewport bool
	vpHeight int
	vpCursor int

	searchable   bool
	searchActive bool
	searchInput  textinput.Model
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
	if t.searchActive {
		t.applySearchFilter()
	} else {
		t.visible = r
	}
	return t
}

func (t *List) WithTitle(title string) *List {
	t.title = title
	return t
}

func (t *List) WithFixedWidth(width int) *List {
	t.fixedWidth = width
	if t.searchable {
		t.searchInput.SetWidth(width)
		t.searchInput.CharLimit = width
	}
	return t
}

func (t *List) WithViewport(height int) *List {
	t.viewport = true
	t.vpHeight = height
	t.vpCursor = 0
	return t
}

func (t *List) WithSearch() *List {
	t.searchable = true
	in := textinput.New()
	in.Prompt = "/"
	in.Placeholder = ""
	if t.fixedWidth > 0 {
		in.SetWidth(t.fixedWidth)
		in.CharLimit = t.fixedWidth
	}
	t.searchInput = in
	return t
}

func (t *List) Filter(filter func(Row) bool) {
	filtered := []Row{}
	for _, r := range t.content {
		if filter(r) {
			filtered = append(filtered, r)
		}
	}
	t.visible = filtered
	t.ResetCursor()
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
	return len(t.visible)
}

func (t *List) Content() []Row {
	return t.visible
}

func (t *List) FocussedRow() Row {
	return t.visible[t.cursor]
}

func (t *List) CursorPos() int {
	return t.cursor
}

func (t *List) SetCursor(idx int) {
	if t.inRange(idx) {
		t.cursor = idx
	}
}

func (t *List) ResetCursor() {
	t.cursor = 0
	t.vpCursor = 0
}

func (t *List) MoveCursor(offset int) tea.Cmd {
	finalOffset := offset

	// skip separator row
	for t.inRange(t.cursor+finalOffset) &&
		(!t.visible[t.cursor+finalOffset].Selectable()) {
		finalOffset += offset
	}

	newCursor := t.cursor + finalOffset

	// exiting list
	if !t.inRange(newCursor) {
		if newCursor < 0 {
			return command.FocusNextElementCmd(command.UpDirection)
		} else {
			return command.FocusNextElementCmd(command.DownDirection)
		}
	}

	// keep cursor in view
	if t.viewport {
		if newCursor < t.vpCursor {
			t.vpCursor = newCursor
		}
		if newCursor >= t.viewportEnd() {
			t.vpCursor += finalOffset
		}
	}

	t.cursor = newCursor
	return nil
}

func NewList(k util.KeyMap, s ListStyles) *List {
	return &List{
		KeyMap:     k,
		Styles:     s,
		fixedWidth: -1,
	}
}

func NewListWithDefaults(km util.KeyMap) *List {
	return &List{
		KeyMap: km,
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
	case tea.KeyPressMsg:
		if t.searchInputFocused() {
			cmd = t.updateSearchInput(msg)
		} else {
			cmd = t.updateRows(msg)
		}
	}
	return t, cmd
}

func (t *List) updateSearchInput(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case key.Matches(msg, searchEscapeKey):
		t.closeSearch()
		return nil
	case key.Matches(msg, t.KeyMap.Down):
		t.searchInput.Blur()
		t.focusFirstRow()
		return nil
	case key.Matches(msg, t.KeyMap.Up):
		return command.FocusNextElementCmd(command.UpDirection)
	default:
		var cmd tea.Cmd
		t.searchInput, cmd = t.searchInput.Update(msg)
		t.applySearchFilter()
		return cmd
	}
}

func (t *List) updateRows(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case t.searchable && key.Matches(msg, t.KeyMap.TextSearch):
		return t.openSearch()
	case key.Matches(msg, t.KeyMap.Up):
		if t.searchActive && t.atTop() {
			return t.searchInput.Focus()
		}
		return t.MoveCursor(-1)
	case key.Matches(msg, t.KeyMap.Down):
		return t.MoveCursor(1)
	case key.Matches(msg, t.KeyMap.Escape):
		if t.searchActive {
			t.closeSearch()
			return nil
		}
		t.focus = false
		return nil
	default:
		if t.cursor < len(t.visible) {
			_, cmd := t.visible[t.cursor].Update(msg)
			return cmd
		}
		return nil
	}
}

func (t *List) openSearch() tea.Cmd {
	t.searchActive = true
	return t.searchInput.Focus()
}

func (t *List) closeSearch() {
	t.searchActive = false
	t.searchInput.Blur()
	t.searchInput.SetValue("")
	t.visible = t.content
	t.ResetCursor()
}

func (t *List) applySearchFilter() {
	t.Filter(SearchFilter(t.searchInput.Value()))
}

func (t *List) searchInputFocused() bool {
	return t.searchActive && t.searchInput.Focused()
}

func (t *List) SearchInputFocused() bool {
	return t.searchInputFocused()
}

// atTop reports whether there is no selectable row above the cursor.
func (t *List) atTop() bool {
	for i := t.cursor - 1; i >= 0; i-- {
		if t.visible[i].Selectable() {
			return false
		}
	}
	return true
}

func (t *List) focusFirstRow() {
	for i, r := range t.visible {
		if r.Selectable() {
			t.cursor = i
			return
		}
	}
	t.cursor = 0
}

func (t *List) View() tea.View {
	var body string
	if t.viewport {
		body = strings.Join(t.toLines()[t.vpCursor:t.viewportEnd()], "\n")
	} else {
		body = t.RenderFullContent()
	}
	if t.searchable && t.searchActive {
		return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, t.renderSearchBar(), body))
	}
	return tea.NewView(body)
}

func (t *List) renderSearchBar() string {
	style := t.Styles.Row
	if t.fixedWidth != -1 {
		style = style.Width(t.fixedWidth)
	}
	return style.Render(t.searchInput.View())
}

func (t *List) toLines() []string {
	return strings.Split(t.RenderFullContent(), "\n")
}

func (t *List) viewportEnd() int {
	height := t.vpHeight
	if t.searchActive {
		height -= searchBarHeight
	}
	return min(len(t.toLines()), t.vpCursor+height)
}

func (t *List) inRange(idx int) bool {
	return idx >= 0 && idx < len(t.visible)
}

func (t *List) RenderFullContent() string {
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

	for i, el := range t.visible {
		elStr := el.View().Content
		var row string
		if t.focus && i == t.cursor && !t.searchInputFocused() {
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
