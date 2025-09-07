package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	app, err := NewApp()
	if err != nil {
		panic(err)
	}
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
