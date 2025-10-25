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
├── LICENSE
├── README.md     // <-- You are here
├── db            // Driver for DuckDB, migration logic + migrations
├── dncapp.go     // Main command handler & coordinator, top-level bubble tea program
├── go.mod
├── go.sum
├── main.go       // Application bootstrap
├── models        // Types reflecting stored data objects & helper types
├── repository    // Interfaces + implementations for data repositories
├── ui            // Screens, editors, other tea models
└── util          // Configs & small utilities
```

## Quick start

Requirements

- Go (recommended 1.25+)
- A terminal that supports alternate screen and UTF-8

Run from the project root:

```bash
# run directly
go run .

# or build and execute
go build -o dnc .
./dnc
```

You can modify the keymap stored at `os.UserConfigDir()/dnc/config.json`. Defaults can be found in `util.DefaultKeyMap()`, but `ctrl+c` should get you out :)

## Data

Data is currently stored in a local DuckDB database using sqlx.

- Location (default): Given by `os.UserConfigDir()` (`~/Library/Application Support/dnc/dnc.db` on macOS)
- Migrations: custom parser, migration files under `db/migrations`.

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

- Currently mostly smaller unit tests. Looking to implement larger integration tests using `teatest`.

## Contributing

- Please open a feature request or a PR if you would like (to contribute) a feature
