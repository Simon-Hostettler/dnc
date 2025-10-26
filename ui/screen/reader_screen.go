package screen

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/ui/command"
	styles "hostettler.dev/dnc/ui/util"
	"hostettler.dev/dnc/ui/viewport"
	"hostettler.dev/dnc/util"
)

var readerHeight = 30

type ReaderScreen struct {
	keymap   util.KeyMap
	viewport *viewport.Viewport
}

func NewReaderScreen(keymap util.KeyMap) *ReaderScreen {
	return &ReaderScreen{
		keymap,
		viewport.NewViewport(keymap, readerHeight, styles.SmallScreenWidth+2),
	}
}

func (s *ReaderScreen) Init() tea.Cmd {
	return s.viewport.Init()
}

func (s *ReaderScreen) StartRead(content string) {
	s.viewport.Reset()
	s.viewport.UpdateContent(
		lipgloss.NewStyle().
			Width(styles.SmallScreenWidth - 2).
			Render(content))
}

func (s *ReaderScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keymap.Escape) || key.Matches(msg, s.keymap.Show):
			cmd = command.SwitchToPrevScreenCmd
		default:
			_, cmd = s.viewport.Update(msg)
		}
	}
	return s, cmd
}

func (s *ReaderScreen) View() string {
	return styles.DefaultBorderStyle.
		Width(styles.SmallScreenWidth + 2).
		Height(readerHeight).
		Render(s.viewport.View())
}

// to fulfill FocusableModel interface
func (s *ReaderScreen) Focus() {}

func (s *ReaderScreen) Blur() {}
