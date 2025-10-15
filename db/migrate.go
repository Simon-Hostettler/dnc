package db

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// migrations holds SQL migration files embedded into the binary.
//
//go:embed migrations/*.sql
var migrations embed.FS

// Marker regexes (simple; assumes markers are on their own lines)
var (
	rxUp   = regexp.MustCompile(`(?s)--\s*\+duckUp\s*(.*?)\n--\s*\+duckDown`)
	rxDown = regexp.MustCompile(`(?s)--\s*\+duckUp.*?--\s*\+duckDown\s*(.*)$`)
)

// migrationFile represents an embedded migration file (meta only until contents read)
type migrationFile struct {
	version int
	name    string
}

// MigrateUp applies all pending migrations. Naming ordered by integer prefix (i.e. 00001_sample.sql)
func MigrateUp(db *sqlx.DB) error {
	if err := ensureMigrationTable(db); err != nil {
		return err
	}
	applied, err := loadAppliedVersions(db)
	if err != nil {
		return err
	}
	list, err := listMigrationFiles()
	if err != nil {
		return err
	}
	for _, mf := range list {
		if applied[mf.version] {
			continue
		}
		upSQL, _, err := loadMigrationSections(mf.name)
		if err != nil {
			return fmt.Errorf("db.MigrateUp: %s: %w", mf.name, err)
		}
		if err := applyUp(db, mf.version, mf.name, upSQL); err != nil {
			return err
		}
	}
	return nil
}

// MigrateDown rolls back all applied migrations in reverse order.
func MigrateDown(db *sqlx.DB) error {
	applied, err := orderedAppliedVersions(db)
	if err != nil {
		return err
	}
	if len(applied) == 0 {
		return nil
	}
	// Build version->filename map once to avoid repeated directory scans.
	migrationIndex, err := buildMigrationIndex()
	if err != nil {
		return err
	}
	for i := len(applied) - 1; i >= 0; i-- { // reverse order
		ver := applied[i]
		name, ok := migrationIndex[ver]
		if !ok {
			return fmt.Errorf("db.MigrateDown: missing file for version %d", ver)
		}
		_, downSQL, err := loadMigrationSections(name)
		if err != nil {
			return fmt.Errorf("db.MigrateDown: %s: %w", name, err)
		}
		if err := applyDown(db, ver, name, downSQL); err != nil {
			return err
		}
	}
	return nil
}

func ensureMigrationTable(db *sqlx.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at TIMESTAMP NOT NULL)`)
	if err != nil {
		return fmt.Errorf("db.ensureMigrationTable: %w", err)
	}
	return nil
}

func loadAppliedVersions(db *sqlx.DB) (map[int]bool, error) {
	rows, err := db.Queryx(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("db.loadAppliedVersions: %w", err)
	}
	defer rows.Close()
	out := map[int]bool{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

func orderedAppliedVersions(db *sqlx.DB) ([]int, error) {
	rows, err := db.Queryx(`SELECT version FROM schema_migrations ORDER BY version ASC`)
	if err != nil {
		return nil, fmt.Errorf("db.orderedAppliedVersions: %w", err)
	}
	defer rows.Close()
	var vs []int
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		vs = append(vs, v)
	}
	return vs, rows.Err()
}

func parseVersion(name string) (int, bool) {
	// Digits until first non-digit.
	for i, r := range name {
		if r < '0' || r > '9' {
			if i == 0 {
				return 0, false
			}
			var v int
			_, err := fmt.Sscanf(name[:i], "%d", &v)
			return v, err == nil
		}
	}
	return 0, false
}

// buildMigrationIndex returns a map[version]filename for all embedded migrations.
func buildMigrationIndex() (map[int]string, error) {
	list, err := listMigrationFiles()
	if err != nil {
		return nil, err
	}
	idx := make(map[int]string, len(list))
	for _, m := range list {
		idx[m.version] = m.name
	}
	return idx, nil
}

// listMigrationFiles enumerates migration files, sorting by version.
func listMigrationFiles() ([]migrationFile, error) {
	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("db.listMigrationFiles: %w", err)
	}
	var list []migrationFile
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ver, ok := parseVersion(e.Name())
		if !ok {
			continue
		}
		list = append(list, migrationFile{version: ver, name: e.Name()})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].version < list[j].version })
	return list, nil
}

// loadMigrationSections returns (upSQL, downSQL) for a file.
func loadMigrationSections(filename string) (string, string, error) {
	content, err := fs.ReadFile(migrations, "migrations/"+filename)
	if err != nil {
		return "", "", fmt.Errorf("read: %w", err)
	}
	raw := string(content)
	up, err := extractUp(raw)
	if err != nil {
		return "", "", err
	}
	down, err := extractDown(raw)
	if err != nil {
		return "", "", err
	}
	return up, down, nil
}

// applyUp executes the up SQL inside a transaction and records the version.
func applyUp(db *sqlx.DB, version int, name, upSQL string) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin up %s: %w", name, err)
	}
	if err = execStatements(tx, upSQL); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("apply up %s: %w", name, err)
	}
	if _, err = tx.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, version, time.Now()); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record version %d: %w", version, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit up %s: %w", name, err)
	}
	return nil
}

// applyDown executes the down SQL inside a transaction and removes the version record.
func applyDown(db *sqlx.DB, version int, name, downSQL string) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin down %s: %w", name, err)
	}
	if err = execStatements(tx, downSQL); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("apply down %s: %w", name, err)
	}
	if _, err = tx.Exec(`DELETE FROM schema_migrations WHERE version = ?`, version); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete version %d: %w", version, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit down %s: %w", name, err)
	}
	return nil
}

func extractUp(content string) (string, error) {
	m := rxUp.FindStringSubmatch(content)
	if len(m) < 2 {
		return "", errors.New("no Up section found")
	}
	return strings.TrimSpace(m[1]), nil
}

func extractDown(content string) (string, error) {
	m := rxDown.FindStringSubmatch(content)
	if len(m) < 2 {
		return "", errors.New("no Down section found")
	}
	return strings.TrimSpace(m[1]), nil
}

// execStatements splits on semicolons and executes non-empty statements.
func execStatements(tx *sqlx.Tx, sqlBlob string) error {
	// Naive split; sufficient for simple DDL. Avoid semicolons inside string literals.
	for _, stmt := range strings.Split(sqlBlob, ";") {
		s := strings.TrimSpace(stmt)
		if s == "" {
			continue
		}
		if _, err := tx.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
