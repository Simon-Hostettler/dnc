package editor

import (
	tea "github.com/charmbracelet/bubbletea"
)

type EditValueMsg struct {
	Editors []ValueEditor
}

type SwitchToEditorMsg struct {
	Editors []ValueEditor
}

func EditValueCmd(editors []ValueEditor) func() tea.Msg {
	return func() tea.Msg {
		return EditValueMsg{editors}
	}
}

func SwitchToEditorCmd(editors []ValueEditor) func() tea.Msg {
	return func() tea.Msg {
		return SwitchToEditorMsg{editors}
	}
}
