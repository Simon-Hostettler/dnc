package ui

import tea "github.com/charmbracelet/bubbletea"

type ValueEditor[V any] interface {
	Init(KeyMap, string, *V)

	Update(tea.Msg) tea.Cmd

	View() string

	Save() tea.Cmd
}
