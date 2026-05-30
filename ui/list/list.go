package list

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

var DefaultColWidth = 16

type Row interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() tea.View
	Editors() []editor.ValueEditor
	Selectable() bool
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

const noSection = -1

// Flattened row together with the index of the Section it belongs to
// (noSection for the inter-section gap separators).
type entry struct {
	row     Row
	section int
}

type List struct {
	KeyMap util.KeyMap
	Styles ListStyles

	focus        bool
	title        string
	sections     []Section
	sectionStyle SectionStyle
	visible      []entry
	cursor       int
	fixedWidth   int

	viewport viewport
	search   search
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
		KeyMap:     km,
		Styles:     DefaultListStyles(),
		fixedWidth: -1,
	}
}

func (t *List) WithRows(r []Row) *List {
	return t.WithSections([]Section{{Items: r}})
}

func (t *List) WithSections(sections []Section) *List {
	t.sections = sections
	t.refresh()
	return t
}

func (t *List) WithSectionStyle(s SectionStyle) *List {
	t.sectionStyle = s
	t.refresh()
	return t
}

func (t *List) WithTitle(title string) *List {
	t.title = title
	return t
}

func (t *List) WithFixedWidth(width int) *List {
	t.fixedWidth = width
	t.search.setWidth(width)
	t.refresh()
	return t
}

func (t *List) WithViewport(height int) *List {
	t.viewport = viewport{enabled: true, height: height}
	return t
}

func (t *List) WithSearch() *List {
	t.search = newSearch(t.fixedWidth)
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
	return len(t.visible)
}

func (t *List) Content() []Row {
	rows := make([]Row, len(t.visible))
	for i, e := range t.visible {
		rows[i] = e.row
	}
	return rows
}

func (t *List) CursorPos() int {
	return t.cursor
}

func (t *List) SetCursor(idx int) {
	if t.inRange(idx) {
		t.cursor = idx
		t.viewport.scrollTo(t.cursor, t.search.barHeight())
	}
}

func (t *List) resetCursor() {
	t.cursor = 0
	t.viewport.reset()
}

func (t *List) inRange(idx int) bool {
	return idx >= 0 && idx < len(t.visible)
}

func (t *List) moveCursor(offset int) tea.Cmd {
	finalOffset := offset

	for t.inRange(t.cursor+finalOffset) &&
		(!t.visible[t.cursor+finalOffset].row.Selectable()) {
		finalOffset += offset
	}

	newCursor := t.cursor + finalOffset

	if !t.inRange(newCursor) {
		if newCursor < 0 {
			return command.FocusNextElementCmd(command.UpDirection)
		}
		return command.FocusNextElementCmd(command.DownDirection)
	}

	t.cursor = newCursor
	t.viewport.scrollTo(t.cursor, t.search.barHeight())
	return nil
}

func (t *List) Init() tea.Cmd {
	return nil
}

func (t *List) Update(m tea.Msg) (tea.Model, tea.Cmd) {
	if !t.focus {
		return t, nil
	}
	var cmd tea.Cmd
	switch msg := m.(type) {
	case tea.KeyPressMsg:
		if t.SearchInputFocused() {
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
		t.search.blur()
		t.focusFirstRow()
		return nil
	case key.Matches(msg, t.KeyMap.Up):
		return command.FocusNextElementCmd(command.UpDirection)
	default:
		cmd := t.search.update(msg)
		t.refresh()
		t.resetCursor()
		return cmd
	}
}

func (t *List) updateRows(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case t.search.enabled && key.Matches(msg, t.KeyMap.TextSearch):
		return t.openSearch()
	case key.Matches(msg, t.KeyMap.Up):
		if t.search.active && t.atTop() {
			return t.search.open()
		}
		return t.moveCursor(-1)
	case key.Matches(msg, t.KeyMap.Down):
		return t.moveCursor(1)
	case key.Matches(msg, t.KeyMap.Escape):
		if t.search.active {
			t.closeSearch()
			return nil
		}
		t.focus = false
		return nil
	case !t.search.active && key.Matches(msg, t.KeyMap.Append):
		return t.triggerAppender()
	default:
		if t.cursor < len(t.visible) {
			_, cmd := t.visible[t.cursor].row.Update(msg)
			return cmd
		}
		return nil
	}
}

func (t *List) triggerAppender() tea.Cmd {
	if t.cursor < 0 || t.cursor >= len(t.visible) {
		return nil
	}
	secIdx := t.visible[t.cursor].section
	if secIdx < 0 || secIdx >= len(t.sections) {
		return nil
	}
	ap := t.sections[secIdx].Appender
	if ap == nil {
		return nil
	}
	return ap.Trigger()
}

func (t *List) openSearch() tea.Cmd {
	return t.search.open()
}

func (t *List) closeSearch() {
	t.search.close()
	t.refresh()
	t.resetCursor()
}

func (t *List) SearchInputFocused() bool {
	return t.search.focused()
}

func (t *List) atTop() bool {
	for i := t.cursor - 1; i >= 0; i-- {
		if t.visible[i].row.Selectable() {
			return false
		}
	}
	return true
}

func (t *List) focusFirstRow() {
	for i, e := range t.visible {
		if e.row.Selectable() {
			t.cursor = i
			return
		}
	}
	t.cursor = 0
}

func (t *List) refresh() {
	if t.search.filtering() {
		t.visible = t.flatten(t.search.filter())
	} else {
		t.visible = t.flatten(nil)
	}
}

// Materializes the sections into a single row list. A nil keep includes every
// item and the section appenders; a non-nil keep filters items, drops sections
// left empty, and omits appenders (the filtered view has no "add" affordance).
func (t *List) flatten(keep func(Row) bool) []entry {
	includeAppenders := keep == nil

	style := t.sectionStyle
	width := style.SeparatorWidth
	if width <= 0 {
		width = t.fixedWidth - 1
	}
	gap := style.SectionGap != "" && width > 0
	headerSep := style.HeaderSeparator != "" && width > 0

	rows := []entry{}
	for i, sec := range t.sections {
		items := sec.Items
		if keep != nil {
			items = nil
			for _, item := range sec.Items {
				if keep(item) {
					items = append(items, item)
				}
			}
			if len(items) == 0 {
				continue
			}
		}

		if len(rows) > 0 && gap {
			rows = append(rows, entry{NewSeparatorRow(style.SectionGap, width), noSection})
		}
		if sec.Header != nil {
			rows = append(rows, entry{sec.Header, i})
			if headerSep {
				rows = append(rows, entry{NewSeparatorRow(style.HeaderSeparator, width), i})
			}
		}
		for _, item := range items {
			rows = append(rows, entry{item, i})
		}
		if includeAppenders && sec.Appender != nil {
			rows = append(rows, entry{sec.Appender, i})
		}
	}
	return rows
}

func (t *List) View() tea.View {
	reserved := t.search.barHeight()
	start, end := t.viewport.window(len(t.visible), reserved)
	body := t.renderRows(start, end)
	if t.title != "" {
		body = lipgloss.JoinVertical(lipgloss.Center, t.renderTitle(), body)
	}
	if t.search.active {
		body = lipgloss.JoinVertical(lipgloss.Left, t.renderSearchBar(), body)
	}
	return tea.NewView(body)
}

func (t *List) renderRows(start, end int) string {
	rows := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		style := t.Styles.Row
		if t.focus && i == t.cursor && !t.SearchInputFocused() {
			style = t.Styles.Selected
		}
		if t.fixedWidth >= 0 {
			style = style.Width(t.fixedWidth)
		}
		rows = append(rows, style.Render(t.visible[i].row.View().Content))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (t *List) renderTitle() string {
	style := t.Styles.Row
	if t.focus {
		style = t.Styles.Selected
	}
	return style.Render(t.title) + "\n"
}

func (t *List) renderSearchBar() string {
	style := t.Styles.Row
	if t.fixedWidth != -1 {
		style = style.Width(t.fixedWidth)
	}
	return style.Render(t.search.view())
}
