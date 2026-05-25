package list

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/editor"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

var DefaultColWidth = 16

const searchBarHeight = 1

var searchEscapeKey = key.NewBinding(key.WithKeys("esc"))

type Row interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() tea.View
	Editors() []editor.ValueEditor
	Selectable() bool
}

type Searchable interface {
	FilterValue() string
}

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

	focus          bool
	title          string
	sections       []Section
	sectionStyle   SectionStyle
	content        []Row
	visible        []Row
	visibleSection []int
	cursor         int
	fixedWidth     int

	viewport bool
	vpHeight int
	vpCursor int

	searchable   bool
	searchActive bool
	searchInput  textinput.Model
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

func (t *List) Init() tea.Cmd {
	return nil
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

func (t *List) inRange(idx int) bool {
	return idx >= 0 && idx < len(t.visible)
}

func (t *List) MoveCursor(offset int) tea.Cmd {
	finalOffset := offset

	for t.inRange(t.cursor+finalOffset) &&
		(!t.visible[t.cursor+finalOffset].Selectable()) {
		finalOffset += offset
	}

	newCursor := t.cursor + finalOffset

	if !t.inRange(newCursor) {
		if newCursor < 0 {
			return command.FocusNextElementCmd(command.UpDirection)
		}
		return command.FocusNextElementCmd(command.DownDirection)
	}

	if t.viewport {
		if newCursor < t.vpCursor {
			t.vpCursor = newCursor
		}
		if newCursor >= t.viewportEnd(len(t.toLines())) {
			t.vpCursor += finalOffset
		}
	}

	t.cursor = newCursor
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
		t.searchInput.Blur()
		t.focusFirstRow()
		return nil
	case key.Matches(msg, t.KeyMap.Up):
		return command.FocusNextElementCmd(command.UpDirection)
	default:
		var cmd tea.Cmd
		t.searchInput, cmd = t.searchInput.Update(msg)
		t.refresh()
		t.ResetCursor()
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
	case !t.searchActive && key.Matches(msg, t.KeyMap.Append):
		return t.triggerAppender()
	default:
		if t.cursor < len(t.visible) {
			_, cmd := t.visible[t.cursor].Update(msg)
			return cmd
		}
		return nil
	}
}

func (t *List) triggerAppender() tea.Cmd {
	if t.cursor < 0 || t.cursor >= len(t.visibleSection) {
		return nil
	}
	secIdx := t.visibleSection[t.cursor]
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
	t.searchActive = true
	return t.searchInput.Focus()
}

func (t *List) closeSearch() {
	t.searchActive = false
	t.searchInput.Blur()
	t.searchInput.SetValue("")
	t.refresh()
	t.ResetCursor()
}

func (t *List) SearchInputFocused() bool {
	return t.searchActive && t.searchInput.Focused()
}

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

func (t *List) refresh() {
	var contentSection []int
	t.content, contentSection = t.flatten(nil, true)
	term := strings.TrimSpace(t.searchInput.Value())
	if t.searchActive && term != "" {
		t.visible, t.visibleSection = t.flatten(SearchFilter(term), false)
	} else {
		t.visible, t.visibleSection = t.content, contentSection
	}
}

func (t *List) flatten(keep func(Row) bool, includeAppenders bool) ([]Row, []int) {
	style := t.sectionStyle
	width := style.SeparatorWidth
	if width <= 0 {
		width = t.fixedWidth - 1
	}

	type kept struct {
		sectionIdx int
		header     Row
		items      []Row
		appender   *AppenderRow
	}
	survivors := make([]kept, 0, len(t.sections))
	for i, sec := range t.sections {
		var items []Row
		if keep == nil {
			items = sec.Items
		} else {
			for _, item := range sec.Items {
				if keep(item) {
					items = append(items, item)
				}
			}
			if len(items) == 0 {
				continue
			}
		}
		k := kept{sectionIdx: i, header: sec.Header, items: items}
		if includeAppenders {
			k.appender = sec.Appender
		}
		survivors = append(survivors, k)
	}

	rows := []Row{}
	sectionOf := []int{}
	for i, s := range survivors {
		if s.header != nil {
			rows = append(rows, s.header)
			sectionOf = append(sectionOf, s.sectionIdx)
			if style.HeaderSeparator != "" && width > 0 {
				rows = append(rows, NewSeparatorRow(style.HeaderSeparator, width))
				sectionOf = append(sectionOf, s.sectionIdx)
			}
		}
		for _, item := range s.items {
			rows = append(rows, item)
			sectionOf = append(sectionOf, s.sectionIdx)
		}
		if s.appender != nil {
			rows = append(rows, s.appender)
			sectionOf = append(sectionOf, s.sectionIdx)
		}
		if i < len(survivors)-1 && style.SectionGap != "" && width > 0 {
			rows = append(rows, NewSeparatorRow(style.SectionGap, width))
			sectionOf = append(sectionOf, -1)
		}
	}
	return rows, sectionOf
}

func (t *List) View() tea.View {
	body := t.RenderBody()
	if t.viewport {
		lines := strings.Split(body, "\n")
		body = strings.Join(lines[t.vpCursor:t.viewportEnd(len(lines))], "\n")
	}
	if t.title != "" {
		body = lipgloss.JoinVertical(lipgloss.Center, t.renderTitle(), body)
	}
	if t.searchable && t.searchActive {
		body = lipgloss.JoinVertical(lipgloss.Left, t.renderSearchBar(), body)
	}
	return tea.NewView(body)
}

func (t *List) RenderBody() string {
	rows := make([]string, 0, len(t.visible))
	for i, el := range t.visible {
		style := t.Styles.Row
		if t.focus && i == t.cursor && !t.SearchInputFocused() {
			style = t.Styles.Selected
		}
		if t.fixedWidth >= 0 {
			style = style.Width(t.fixedWidth)
		}
		rows = append(rows, style.Render(el.View().Content))
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
	return style.Render(t.searchInput.View())
}

func (t *List) toLines() []string {
	return strings.Split(t.RenderBody(), "\n")
}

func (t *List) viewportEnd(total int) int {
	height := t.vpHeight
	if t.searchActive {
		height -= searchBarHeight
	}
	return min(total, t.vpCursor+height)
}
