package screen

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"hostettler.dev/dnc/command"
	"hostettler.dev/dnc/ui/styles"
	"hostettler.dev/dnc/util"
)

const (
	confirmationScreenHeight = 10
	confirmationScreenWidth  = 40
)

type ConfirmationScreen struct {
	keymap       util.KeyMap
	callback     func() tea.Cmd
	confirmation bool
}

func NewConfirmationScreen(keymap util.KeyMap) *ConfirmationScreen {
	return &ConfirmationScreen{
		keymap,
		func() tea.Cmd { return nil },
		false,
	}
}

func (s *ConfirmationScreen) Init() tea.Cmd {
	return nil
}

func (s *ConfirmationScreen) LaunchConfirmation(callback func() tea.Cmd) {
	s.callback = callback
}

func (s *ConfirmationScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keymap.Enter):
			if s.confirmation {
				return s, tea.Batch(s.callback(), command.SwitchToPrevScreenCmd)
			} else {
				return s, command.SwitchToPrevScreenCmd
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
	dialogue := styles.DefaultTextStyle.
		Height(confirmationScreenHeight/2 - 1).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render("Are you sure?")

	confirmButton := styles.RenderItem(s.confirmation, "[ Yes ]")
	declineButton := styles.RenderItem(!s.confirmation, "[ No ]")

	buttons := lipgloss.PlaceVertical(
		confirmationScreenHeight/2-1,
		lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Center,
			lipgloss.PlaceHorizontal(confirmationScreenWidth/2-1, lipgloss.Center, confirmButton),
			lipgloss.PlaceHorizontal(confirmationScreenWidth/2-1, lipgloss.Center, declineButton),
		),
	)

	content := lipgloss.JoinVertical(lipgloss.Center, dialogue, buttons)

	return styles.DefaultBorderStyle.
		Render(content)
}

// to fulfill FocusableModel interface
func (s *ConfirmationScreen) Focus() {}

func (s *ConfirmationScreen) Blur() {}
