```
______ _   _ _____
|  _  \ \ | /  __ \
| | | |  \| | /  \/
| | | | . ` | |
| |/ /| |\  | \__/\
|___/ \_| \_/\____/
```
[![Go](https://github.com/Simon-Hostettler/dnc/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/Simon-Hostettler/dnc/actions/workflows/go.yml)

A small terminal-based TTRPG character manager built with Bubble Tea, Lip Gloss and DuckDB. It provides a TUI to create and edit various character data (stats, spells, items, etc.).

This repository is organized as a single Go module (`hostettler.dev/dnc`) with the following rough layout:

```
├── README.md     // <-- You are here
├── config.go     // Basic configurations - in code only for the moment
├── db            // Driver for DuckDB, migration logic + migrations
├── dncapp.go     // Main command handler & coordinator, top-level bubble tea program
├── go.mod        // Deps
├── go.sum
├── main.go       // Application bootsrap
├── models        // Types reflecting stored data objects & helper types
├── repository    // Interfaces + implementations for data repositories
└── ui            // Screens, editors, other tea models
```

## Quick start

Requirements

- Go (recommended 1.20+)
- A terminal that supports alternate screen and UTF-8

Run from the project root:

```bash
# run directly
go run .

# or build and execute
go build -o dnc .
./dnc
```

The exact keymap for the moment is defined in `util.DefaultKeyMap()`, but `ctrl+c` should get you out :)

## Data

Data is currently stored in a local DuckDB database using sqlx.

- Location (default): `~/Library/Application Support/dnc/dnc.db` (on macOS via `os.UserConfigDir()`)
- Migrations: embedded with `pressly/goose` under `db/migrations`.

To add a migration:

1. Create a new file in `db/migrations` named like `0002_add_foo.sql`.
2. Include sections:
```
   -- +duckUp
   -- SQL to apply
   -- +duckDown
   -- SQL to roll back
```

Migrations will be applied automatically at startup

## Testing

- TODO :)

## Contributing

- Please open a feature request or a PR if you would like (to contribute) a feature
