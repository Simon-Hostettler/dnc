package main

import (
	"flag"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"hostettler.dev/dnc/util"
)

func main() {
	demo := flag.Bool("demo", false, "start with a temporary demo database")
	flag.Parse()

	config, cleanup, err := util.GetConfig(util.DefaultConfigDir(), *demo)
	if err != nil {
		log.Fatal(err)
	}

	app, err := NewApp(config, cleanup)
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
