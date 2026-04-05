package main

import (
	"flag"
	"log"
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"hostettler.dev/dnc/util"
)

func main() {
	demo := flag.Bool("demo", false, "start with a temporary demo database")
	flag.Parse()

	cfgDir := util.DefaultConfigDir()
	logCleanup, err := util.InitLogger(cfgDir, 5*1024*1024)
	if err != nil {
		log.Fatal("failed to initialise log file: ", err)
	}
	defer logCleanup()

	slog.Info("dnc starting", "demo", *demo)

	config, cleanup, err := util.GetConfig(cfgDir, *demo)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		log.Fatal(err)
	}

	app, err := NewApp(config, cleanup)
	if err != nil {
		slog.Error("failed to initialise app", "error", err)
		log.Fatal(err)
	}
	defer app.Close()

	p := tea.NewProgram(app)

	if _, err := p.Run(); err != nil {
		slog.Error("program exited with error", "error", err)
		log.Fatal(err)
	}

	slog.Info("dnc exited cleanly")
}
