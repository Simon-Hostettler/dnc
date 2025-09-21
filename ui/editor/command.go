package editor

import (
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/models"
	"hostettler.dev/dnc/ui/command"
)

type EditValueMsg struct {
	Editors []ValueEditor
}

type SwitchToEditorMsg struct {
	Originator command.ScreenIndex
	Character  *models.Character
	Editors    []ValueEditor
}

func EditValueCmd(editors []ValueEditor) func() tea.Msg {
	return func() tea.Msg {
		return EditValueMsg{editors}
	}
}

func SwitchToEditorCmd(caller command.ScreenIndex, character *models.Character, editors []ValueEditor) func() tea.Msg {
	return func() tea.Msg {
		return SwitchToEditorMsg{caller, character, editors}
	}
}
