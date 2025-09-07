package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/ui"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	p := tea.NewProgram(ui.NewTitleScreen(cfg.CharacterDir))
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
