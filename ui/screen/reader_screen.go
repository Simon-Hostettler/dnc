package screen

import (
	"charm.land/bubbles/v2/key"
	ti "charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/ui/textinput"
	"hostettler.dev/dnc/ui/viewport"
	"hostettler.dev/dnc/util"
)

var (
	ReaderHeight     = 30
	readerInnerWidth = styles.SmallScreenWidth - 2
)

var matchStyle = lipgloss.NewStyle().Background(styles.HighlightColor)

type ReaderScreen struct {
	keymap      util.KeyMap
	viewport    *viewport.Viewport
	searchField *textinput.TextInput
	searchMode  bool
}

func NewReaderScreen(keymap util.KeyMap) *ReaderScreen {
	sf := ti.New()
	sf.SetWidth(readerInnerWidth)
	sf.CharLimit = readerInnerWidth
	sf.Placeholder = ""
	sf.Prompt = "/"

	return &ReaderScreen{
		keymap:      keymap,
		viewport:    viewport.NewViewport(keymap, ReaderHeight, readerInnerWidth),
		searchField: textinput.New(sf),
		searchMode:  false,
	}
}

func (s *ReaderScreen) Init() tea.Cmd {
	return s.viewport.Init()
}

func (s *ReaderScreen) StartRead(content string) {
	s.searchMode = false
	s.searchField.SetValue("")
	s.viewport.Reset()
	s.viewport.UpdateContent(content)
}

func (s *ReaderScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if s.searchMode {
			switch {
			case key.Matches(msg, s.keymap.Escape):
				s.searchField.SetValue("")
				s.viewport.ClearHighlight()
				s.searchField.Blur()
				s.searchMode = false
			case key.Matches(msg, s.keymap.Enter):
				s.highlightText(s.searchField.Value())
				s.searchField.Blur()
				s.searchMode = false
			default:
				_, cmd = s.searchField.Update(msg)
			}
		} else {
			switch {
			case key.Matches(msg, s.keymap.Escape) || key.Matches(msg, s.keymap.Show):
				cmd = command.SwitchToPrevScreenCmd
			case key.Matches(msg, s.keymap.TextSearch):
				s.searchMode = true
				s.searchField.Focus()
			default:
				_, cmd = s.viewport.Update(msg)
			}
		}
	}
	return s, cmd
}

func (s *ReaderScreen) View() tea.View {
	borderHeight := ReaderHeight + 2
	var inner string
	if s.searchMode {
		borderHeight = ReaderHeight + 3
		inner = lipgloss.JoinVertical(lipgloss.Left,
			s.viewport.View().Content,
			s.searchField.View().Content,
		)
	} else {
		inner = s.viewport.View().Content
	}
	return tea.NewView(styles.DefaultBorderStyle.
		Width(readerInnerWidth + 4).
		Height(borderHeight).
		Align(lipgloss.Left).
		Render(inner))
}

func (s *ReaderScreen) highlightText(match string) {
	if match != "" {
		s.viewport.SetHighlight(match, matchStyle, styles.DefaultTextStyle)
	} else {
		s.viewport.ClearHighlight()
	}
}

// to fulfill FocusableModel interface
func (s *ReaderScreen) Focus() {}

func (s *ReaderScreen) Blur() {}
