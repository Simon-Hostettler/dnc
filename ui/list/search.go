package list

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

const searchBarHeight = 1

var searchEscapeKey = key.NewBinding(key.WithKeys("esc"))

type Searchable interface {
	FilterValue() string
}

// Predicate matching rows whose FilterValue contains term (case-insensitive)
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

type search struct {
	enabled bool
	active  bool
	input   textinput.Model
}

func newSearch(fixedWidth int) search {
	in := textinput.New()
	in.Prompt = "/"
	in.Placeholder = ""
	if fixedWidth > 0 {
		in.SetWidth(fixedWidth)
		in.CharLimit = fixedWidth
	}
	return search{enabled: true, input: in}
}

func (s *search) setWidth(w int) {
	if !s.enabled {
		return
	}
	s.input.SetWidth(w)
	s.input.CharLimit = w
}

func (s *search) open() tea.Cmd {
	s.active = true
	return s.input.Focus()
}

func (s *search) close() {
	s.active = false
	s.input.Blur()
	s.input.SetValue("")
}

func (s *search) blur() {
	s.input.Blur()
}

func (s *search) focused() bool {
	return s.active && s.input.Focused()
}

func (s *search) term() string {
	return strings.TrimSpace(s.input.Value())
}

func (s *search) filtering() bool {
	return s.active && s.term() != ""
}

func (s *search) filter() func(Row) bool {
	return SearchFilter(s.term())
}

func (s *search) barHeight() int {
	if s.active {
		return searchBarHeight
	}
	return 0
}

func (s *search) update(msg tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return cmd
}

func (s *search) view() string {
	return s.input.View()
}
