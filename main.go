package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"hostettler.dev/dnc/util"
)

func main() {
	demo := flag.Bool("demo", false, "start with a temporary demo database")
	backup := flag.String("backup", "", "copy database to specified file path")
	restore := flag.String("restore", "", "overwrite database with specified file path")
	flag.Parse()

	cfgDir := util.DefaultConfigDir()
	logCleanup, err := util.InitLogger(cfgDir, 5*1024*1024)
	if err != nil {
		log.Fatal("failed to initialise log file: ", err)
	}
	defer logCleanup()

	dbPath := util.DefaultConfig(cfgDir).DatabasePath

	if *backup != "" {
		slog.Info("backup requested", "src", dbPath, "dst", *backup)
		if err := util.CopyFile(dbPath, *backup); err != nil {
			log.Fatal("backup failed: ", err)
		}
		fmt.Printf("Database backed up to %s\n", *backup)
		os.Exit(0)
	}

	if *restore != "" {
		confirmation_string := "I am aware that this action overwrites all my current data"

		fmt.Println("WARNING: This will overwrite all current data in your database.")
		fmt.Printf("Type the following to confirm: %s\n> ", confirmation_string)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.TrimRight(scanner.Text(), "\r\n")
		if input != confirmation_string {
			fmt.Println("Aborted.")
			os.Exit(1)
		}
		slog.Info("restore requested", "src", *restore, "dst", dbPath)
		if err := util.CopyFile(*restore, dbPath); err != nil {
			log.Fatal("restore failed: ", err)
		}
		fmt.Printf("Database restored from %s\n", *restore)
		os.Exit(0)
	}

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
