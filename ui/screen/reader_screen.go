package screen

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui/command"
	"hostettler.dev/dnc/ui/util"
)

var readerHeight = 30

type ReaderScreen struct {
	keymap    util.KeyMap
	character *models.Character
	cursor    int
	content   string
}

func NewReaderScreen(keymap util.KeyMap) *ReaderScreen {
	return &ReaderScreen{keymap, nil, 0, ""}
}

func (s *ReaderScreen) Init() tea.Cmd {
	return nil
}

func (s *ReaderScreen) StartRead(content string) {
	s.cursor = 0
	s.content = content
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
		case key.Matches(msg, s.keymap.Escape):
			cmd = command.SwitchToPrevScreenCmd
		}
	}
	return s, cmd
}

func (s *ReaderScreen) View() string {
	viewableContent := strings.Join(s.contentToLines()[s.cursor:s.cursorEnd()], "\n")

	return util.DefaultBorderStyle.
		Width(util.SmallScreenWidth).
		Render(viewableContent)
}

func (s *ReaderScreen) contentToLines() []string {
	return strings.Split(s.content, "\n")
}

func (s *ReaderScreen) cursorEnd() int {
	return min(len(s.contentToLines()), s.cursor+readerHeight)
}

// to fulfill FocusableModel interface
func (s *ReaderScreen) Focus() {}

func (s *ReaderScreen) Blur() {}
