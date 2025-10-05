package screen

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/ui/command"
	"hostettler.dev/dnc/ui/util"
)

const (
	confirmationScreenHeight = 10
	confirmationScreenWidth  = 40
)

type ConfirmationScreen struct {
	keymap       util.KeyMap
	prevScreen   command.ScreenIndex
	callback     func() tea.Cmd
	confirmation bool
}

func NewConfirmationScreen(keymap util.KeyMap) *ConfirmationScreen {
	return &ConfirmationScreen{
		keymap,
		command.ConfirmationScreenIndex,
		func() tea.Cmd { return nil },
		false,
	}
}

func (s *ConfirmationScreen) Init() tea.Cmd {
	return nil
}

func (s *ConfirmationScreen) LaunchConfirmation(prevScreen command.ScreenIndex, callback func() tea.Cmd) {
	s.prevScreen = prevScreen
	s.callback = callback
}

func (s *ConfirmationScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keymap.Enter):
			if s.confirmation {
				return s, tea.Batch(s.callback(), command.SwitchScreenCmd(s.prevScreen))
			} else {
				return s, command.SwitchScreenCmd(s.prevScreen)
			}
		case key.Matches(msg, s.keymap.Left) && !s.confirmation:
			s.confirmation = true
		case key.Matches(msg, s.keymap.Right) && s.confirmation:
			s.confirmation = false
		}
	}
	return s, cmd
}

func (s *ConfirmationScreen) View() string {
	dialogue := util.DefaultTextStyle.
		Height(confirmationScreenHeight/2 - 1).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render("Are you sure?")

	confirmButton := util.RenderItem(s.confirmation, "[ Yes ]")
	declineButton := util.RenderItem(!s.confirmation, "[ No ]")

	buttons := lipgloss.PlaceVertical(
		confirmationScreenHeight/2-1,
		lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Center,
			lipgloss.PlaceHorizontal(confirmationScreenWidth/2-1, lipgloss.Center, confirmButton),
			lipgloss.PlaceHorizontal(confirmationScreenWidth/2-1, lipgloss.Center, declineButton),
		),
	)

	content := lipgloss.JoinVertical(lipgloss.Center, dialogue, buttons)

	return util.DefaultBorderStyle.
		Render(content)
}

// to fulfill FocusableModel interface
func (s *ConfirmationScreen) Focus() {}

func (s *ConfirmationScreen) Blur() {}
