package ui

import tea "github.com/charmbracelet/bubbletea"

type ValueEditor interface {
	Init(string, interface{}, KeyMap)

	Update(tea.Msg) tea.Cmd

	View() string

	Save() tea.Cmd
}
