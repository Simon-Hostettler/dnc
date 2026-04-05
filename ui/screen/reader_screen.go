package screen

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/ui/viewport"
	"hostettler.dev/dnc/util"
)

var ReaderHeight = 30

type ReaderScreen struct {
	keymap   util.KeyMap
	viewport *viewport.Viewport
}

func NewReaderScreen(keymap util.KeyMap) *ReaderScreen {
	return &ReaderScreen{
		keymap,
		viewport.NewViewport(keymap, ReaderHeight, styles.SmallScreenWidth-2),
	}
}

func (s *ReaderScreen) Init() tea.Cmd {
	return s.viewport.Init()
}

func (s *ReaderScreen) StartRead(content string) {
	s.viewport.Reset()
	s.viewport.UpdateContent(content)
}

func (s *ReaderScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, s.keymap.Escape) || key.Matches(msg, s.keymap.Show):
			cmd = command.SwitchToPrevScreenCmd
		default:
			_, cmd = s.viewport.Update(msg)
		}
	}
	return s, cmd
}

func (s *ReaderScreen) View() tea.View {
	return tea.NewView(styles.DefaultBorderStyle.
		Width(styles.SmallScreenWidth + 2).
		Height(ReaderHeight + 2).
		Align(lipgloss.Left).
		Render(s.viewport.View().Content))
}

// to fulfill FocusableModel interface
func (s *ReaderScreen) Focus() {}

func (s *ReaderScreen) Blur() {}
