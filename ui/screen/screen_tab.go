package screen

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/ui/command"
	styles "hostettler.dev/dnc/ui/util"
	"hostettler.dev/dnc/util"
)

var (
	tabWidth  = 11
	tabHeight = 3
)

type ScreenTab struct {
	keymap      util.KeyMap
	name        string
	screenIndex command.ScreenIndex
	focus       bool
}

func NewScreenTab(keymap util.KeyMap, name string, idx command.ScreenIndex, focus bool) *ScreenTab {
	return &ScreenTab{keymap, name, idx, focus}
}

func (s *ScreenTab) Init() tea.Cmd {
	return nil
}

func (s *ScreenTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, s.keymap.Enter) {
			cmd = command.SwitchScreenCmd(s.screenIndex)
		}
	}
	return s, cmd
}

func (s *ScreenTab) View() string {
	name := s.name
	if s.focus {
		name = styles.ItemStyleSelected.Render(name)
	} else {
		name = styles.ItemStyleDefault.Render(name)
	}
	return styles.DefaultBorderStyle.UnsetPadding().
		AlignVertical(lipgloss.Center).
		Width(tabWidth).
		Height(tabHeight).
		Render(name)
}

func (s *ScreenTab) Focus() {
	s.focus = true
}

func (s *ScreenTab) Blur() {
	s.focus = false
}
