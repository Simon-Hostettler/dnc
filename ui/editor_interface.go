package ui

import tea "github.com/charmbracelet/bubbletea"

type ValueEditor interface {
	Init(KeyMap, string, interface{})

	Update(tea.Msg) tea.Cmd

	View() string

	Save() tea.Cmd
}
