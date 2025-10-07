package editor

import (
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/models"
)

type EditValueMsg struct {
	Editors []ValueEditor
}

type SwitchToEditorMsg struct {
	Character *models.Character
	Editors   []ValueEditor
}

func EditValueCmd(editors []ValueEditor) func() tea.Msg {
	return func() tea.Msg {
		return EditValueMsg{editors}
	}
}

func SwitchToEditorCmd(character *models.Character, editors []ValueEditor) func() tea.Msg {
	return func() tea.Msg {
		return SwitchToEditorMsg{character, editors}
	}
}
