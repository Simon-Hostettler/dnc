package screen

import (
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/ui/command"
)

func UpdateFilesCmd(t *TitleScreen) func() tea.Msg {
	return func() tea.Msg {
		t.UpdateFiles()
		return command.DataOpMsg{
			Op:      command.DataUpdate,
			Success: true,
		}
	}
}
