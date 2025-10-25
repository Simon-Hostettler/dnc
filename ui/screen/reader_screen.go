package screen

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/ui/command"
	styles "hostettler.dev/dnc/ui/util"
	"hostettler.dev/dnc/util"
)

var readerHeight = 30

type ReaderScreen struct {
	keymap  util.KeyMap
	cursor  int
	content string
}

func NewReaderScreen(keymap util.KeyMap) *ReaderScreen {
	return &ReaderScreen{keymap, 0, ""}
}

func (s *ReaderScreen) Init() tea.Cmd {
	return nil
}

func (s *ReaderScreen) StartRead(content string) {
	s.cursor = 0
	s.content = lipgloss.NewStyle().
		Width(styles.SmallScreenWidth - 2).
		Render(content)
}

func (s *ReaderScreen) MoveCursor(offset int) {
	newCursor := s.cursor + offset

	if newCursor >= 0 && newCursor+readerHeight < len(s.contentToLines()) {
		s.cursor = newCursor
	}
}

func (s *ReaderScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keymap.Up):
			s.MoveCursor(-1)
		case key.Matches(msg, s.keymap.Down, s.keymap.Enter):
			s.MoveCursor(1)
		case key.Matches(msg, s.keymap.Escape) || key.Matches(msg, s.keymap.Show):
			cmd = command.SwitchToPrevScreenCmd
		}
	}
	return s, cmd
}

func (s *ReaderScreen) View() string {
	viewableContent := strings.Join(s.contentToLines()[s.cursor:s.cursorEnd()], "\n")

	return styles.DefaultBorderStyle.
		Width(styles.SmallScreenWidth + 2).
		Height(readerHeight).
		Render(lipgloss.PlaceVertical(readerHeight, lipgloss.Left, viewableContent))
}

func (s *ReaderScreen) contentToLines() []string {
	return strings.Split(s.content, "\n")
}

func (s *ReaderScreen) cursorEnd() int {
	return min(len(s.contentToLines()), s.cursor+readerHeight+1)
}

// to fulfill FocusableModel interface
func (s *ReaderScreen) Focus() {}

func (s *ReaderScreen) Blur() {}
